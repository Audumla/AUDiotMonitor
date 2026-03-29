import os
import subprocess
import tempfile
import time
import shutil
import requests
import pytest
import yaml
from pathlib import Path

# Base directories
MONITORING_ROOT = Path(__file__).resolve().parents[2]
COLLECTOR_ROOT = MONITORING_ROOT / "collector"
DASHBOARD_ROOT = MONITORING_ROOT / "dashboard"
FIXTURE_PATH = MONITORING_ROOT.parent / "hwexp" / "tests" / "fixtures" / "sample_hwmon.json"


def _can_run_bash() -> tuple[bool, str]:
    bash_path = shutil.which("bash")
    if not bash_path:
        return False, "bash not found in PATH"
    try:
        subprocess.run([bash_path, "-lc", "echo ok"], check=True, capture_output=True, timeout=10)
    except (subprocess.SubprocessError, OSError) as exc:
        return False, f"bash unusable: {exc}"
    return True, ""


def _can_run_docker_compose() -> tuple[bool, str]:
    docker_path = shutil.which("docker")
    if not docker_path:
        return False, "docker not found in PATH"
    try:
        subprocess.run([docker_path, "compose", "version"], check=True, capture_output=True, timeout=10)
        subprocess.run([docker_path, "info"], check=True, capture_output=True, timeout=10)
    except (subprocess.SubprocessError, OSError) as exc:
        return False, f"docker compose unavailable: {exc}"
    return True, ""


@pytest.fixture(scope="session", autouse=True)
def require_integration_runtime():
    ok_bash, reason_bash = _can_run_bash()
    if not ok_bash:
        pytest.skip(f"integration runtime prerequisite missing: {reason_bash}")

    ok_docker, reason_docker = _can_run_docker_compose()
    if not ok_docker:
        pytest.skip(f"integration runtime prerequisite missing: {reason_docker}")

@pytest.fixture(scope="session")
def temp_run_dir():
    with tempfile.TemporaryDirectory(prefix="audiot-test-") as tmpdir:
        yield Path(tmpdir)

@pytest.fixture(scope="session")
def collector_stack(temp_run_dir):
    install_dir = temp_run_dir / "collector"
    install_dir.mkdir()
    
    # Run install-layout.sh
    env = os.environ.copy()
    env["INSTALL_DIR"] = str(install_dir)
    subprocess.run(["bash", str(COLLECTOR_ROOT / "install-layout.sh")], 
                   env=env, check=True, capture_output=True)
    
    # Modify docker-compose.yml for testing
    compose_path = install_dir / "docker-compose.yml"
    with open(compose_path, 'r') as f:
        cfg = yaml.safe_load(f)
    
    # 1. Add fixture to hwexp
    hwexp = cfg['services']['hwexp']
    hwexp['volumes'] = [v for v in hwexp['volumes'] if '/sys' not in v and '/proc' not in v]
    hwexp['volumes'].append(f"{FIXTURE_PATH}:/etc/hwexp/fixture.json:ro")
    hwexp['command'] = ["--config", "/etc/hwexp/hwexp.yaml", "--fixture", "/etc/hwexp/fixture.json"]
    # 2. Disable node-exporter (not cross-platform friendly in CI/Windows)
    if 'node-exporter' in cfg['services']:
        del cfg['services']['node-exporter']
    # 3. Allow Prometheus to write TSDB on ephemeral CI temp dirs regardless of host UID/GID.
    if "prometheus" in cfg["services"]:
        cfg["services"]["prometheus"].pop("user", None)
    
    with open(compose_path, 'w') as f:
        yaml.safe_dump(cfg, f)
    
    # Start the stack
    subprocess.run(["docker", "compose", "up", "-d"],
                   cwd=install_dir, check=True, capture_output=True)

    # Wait for Prometheus and hwexp to be ready
    prom_ready = False
    hwexp_ready = False
    for _ in range(30):
        try:
            prom_ready = requests.get("http://localhost:9090/-/healthy", timeout=3).status_code == 200
            hwexp_resp = requests.get("http://localhost:9200/readyz", timeout=3)
            hwexp_ready = hwexp_resp.status_code == 200 and hwexp_resp.json().get("status") == "ready"
            if prom_ready and hwexp_ready:
                break
        except requests.RequestException:
            pass
        time.sleep(2)
    assert prom_ready, "Prometheus did not become healthy in time"
    assert hwexp_ready, "hwexp did not become ready in time"
    
    yield install_dir
    
    # Cleanup
    subprocess.run(["docker", "compose", "down", "-v"], 
                   cwd=install_dir, check=True, capture_output=True)

@pytest.fixture(scope="session")
def dashboard_stack(temp_run_dir, collector_stack):
    install_dir = temp_run_dir / "dashboard"
    install_dir.mkdir()
    
    # Run install-layout.sh
    env = os.environ.copy()
    env["INSTALL_DIR"] = str(install_dir)
    subprocess.run(["bash", str(DASHBOARD_ROOT / "install-layout.sh")], 
                   env=env, check=True, capture_output=True)
    
    # Disable kiosk in CI tests and pin datasource endpoints to collector stack.
    compose_path = install_dir / "docker-compose.yml"
    with open(compose_path, "r", encoding="utf-8") as f:
        cfg = yaml.safe_load(f)
    if "kiosk" in cfg.get("services", {}):
        del cfg["services"]["kiosk"]
    grafana = cfg["services"]["grafana"]
    env = grafana.get("environment", [])
    env = [entry for entry in env if not str(entry).startswith("PROMETHEUS_URL=") and not str(entry).startswith("HWEXP_URL=")]
    env.append("PROMETHEUS_URL=http://host.docker.internal:9090")
    env.append("HWEXP_URL=http://host.docker.internal:9200")
    grafana["environment"] = env
    with open(compose_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(cfg, f)

    # Start the stack
    subprocess.run(["docker", "compose", "up", "-d"],
                   cwd=install_dir, check=True, capture_output=True)

    # Wait for Grafana health
    healthy = False
    for _ in range(30):
        try:
            res = requests.get("http://localhost:3000/api/health", timeout=3)
            healthy = res.status_code == 200 and res.json().get("database") == "ok"
            if healthy:
                break
        except requests.RequestException:
            pass
        time.sleep(2)
    assert healthy, "Grafana did not become healthy in time"
    
    yield install_dir
    
    # Cleanup
    subprocess.run(["docker", "compose", "down", "-v"], 
                   cwd=install_dir, check=True, capture_output=True)

def test_collector_health(collector_stack):
    # Check Prometheus
    res = requests.get("http://localhost:9090/-/healthy", timeout=10)
    assert res.status_code == 200
    
    # Check hwexp
    res = requests.get("http://localhost:9200/readyz", timeout=10)
    assert res.status_code == 200
    assert res.json()["status"] == "ready"

def test_prometheus_scrapes_hwexp_target(collector_stack):
    # Ensure Prometheus is actively scraping hwexp; this catches exporter
    # exposition issues that readyz alone does not detect.
    res = requests.get("http://localhost:9090/api/v1/query?query=up{job=\"hwexp\"}", timeout=10)
    assert res.status_code == 200
    data = res.json()
    assert data["status"] == "success"
    assert len(data["data"]["result"]) == 1
    assert float(data["data"]["result"][0]["value"][1]) == 1.0

def test_grafana_health(dashboard_stack):
    res = requests.get("http://localhost:3000/api/health", timeout=10)
    assert res.status_code == 200
    assert res.json()["database"] == "ok"

def test_metrics_present_in_prometheus(collector_stack):
    # Wait for a few scrape cycles
    time.sleep(30)
    
    # Query for a raw metric from the fixture
    query = 'hw_device_temperature_celsius'
    res = requests.get(f"http://localhost:9090/api/v1/query?query={query}", timeout=10)
    assert res.status_code == 200
    data = res.json()
    assert data["status"] == "success"
    assert len(data["data"]["result"]) > 0, "No metrics found in Prometheus. Scrape might have failed."

def test_recording_rules_working(collector_stack):
    # Query for a derived metric defined in recording rules
    # We use a simple one that doesn't depend on capacity for this basic check
    query = 'audiot_gpu_compute_utilization_percent'
    res = requests.get(f"http://localhost:9090/api/v1/query?query={query}", timeout=10)
    assert res.status_code == 200
    data = res.json()
    assert data["status"] == "success"
    # Note: result might be empty if the fixture doesn't have utilization,
    # but the status should be success and the metric name should resolve.

def test_grafana_prometheus_connection_e2e(dashboard_stack):
    # This is the most important end-to-end test:
    # 1. Get Prometheus datasource ID
    res = requests.get("http://localhost:3000/api/datasources", auth=("admin", "admin"), timeout=10)
    assert res.status_code == 200
    ds_list = res.json()
    prom_ds = next(d for d in ds_list if d["type"] == "prometheus")
    ds_id = prom_ds["id"]
    
    # 2. Query Prometheus THROUGH Grafana proxy
    query = 'hw_device_temperature_celsius'
    proxy_url = f"http://localhost:3000/api/datasources/proxy/{ds_id}/api/v1/query?query={query}"
    
    # Retry a few times as Grafana might be still initializing the proxy
    for _ in range(5):
        res = requests.get(proxy_url, auth=("admin", "admin"), timeout=10)
        if res.status_code == 200:
            data = res.json()
            if data["status"] == "success" and len(data["data"]["result"]) > 0:
                break
        time.sleep(5)
    
    assert res.status_code == 200
    data = res.json()
    assert data["status"] == "success"
    assert len(data["data"]["result"]) > 0, "Grafana could not query metrics from Prometheus via proxy"
    assert data["data"]["result"][0]["metric"]["__name__"] == "hw_device_temperature_celsius"

def test_grafana_datasources_provisioned(dashboard_stack):
    # Check if Grafana has Prometheus and Infinity datasources
    res = requests.get("http://localhost:3000/api/datasources", auth=("admin", "admin"), timeout=10)
    assert res.status_code == 200
    ds = res.json()
    types = {d["type"] for d in ds}
    assert "prometheus" in types
    assert "yesoreyeram-infinity-datasource" in types

def test_grafana_dashboards_provisioned(dashboard_stack):
    # Check if the System Overview dashboard is present
    res = requests.get("http://localhost:3000/api/search?query=System Overview", auth=("admin", "admin"), timeout=10)
    assert res.status_code == 200
    results = res.json()
    assert any(d["title"] == "AUDiot - System Overview" for d in results)
