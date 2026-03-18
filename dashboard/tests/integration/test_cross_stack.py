import pytest
import requests
import subprocess
import time
import os

# Helper to run docker compose
def run_compose(dir_path, cmd):
    cmd_str = f"docker compose {' '.join(cmd)}"
    res = subprocess.run(cmd_str, cwd=dir_path, capture_output=True, text=True, shell=True, env=os.environ)
    if res.returncode != 0:
        print(f"Docker Compose error in {dir_path}: {res.stderr}")
    return res

import yaml

@pytest.fixture(scope="session", autouse=True)
def manage_stacks():
    server_dir = os.path.abspath("AUDiotMonitor/dashboard/server")
    display_dir = os.path.abspath("AUDiotMonitor/dashboard/display")
    
    # 1. Start Server Stack with mock fixture
    compose_path = os.path.join(server_dir, "docker-compose.yml")
    with open(compose_path, 'r') as f:
        original_compose = f.read()
    
    # Use yaml library to modify
    server_cfg = yaml.safe_load(original_compose)
    hwexp = server_cfg['services']['hwexp']
    # Add volume
    if 'volumes' not in hwexp: hwexp['volumes'] = []
    hwexp['volumes'].append('../../hwexp/tests/fixtures/sample_hwmon.json:/etc/hwexp/fixture.json:ro')
    # Change command
    hwexp['command'] = ["--config", "/etc/hwexp/hwexp.yaml", "--fixture", "/etc/hwexp/fixture.json"]
    
    with open(compose_path, 'w') as f:
        yaml.safe_dump(server_cfg, f)
    
    # 2. Start Display Stack
    display_compose_path = os.path.join(display_dir, "docker-compose.yml")
    with open(display_compose_path, 'r') as f:
        original_display_compose = f.read()
    
    # Modify provisioning file
    prov_file = os.path.join(display_dir, "config/grafana/provisioning/datasources/prometheus.yaml")
    with open(prov_file, 'r') as f:
        original_prov = f.read()
    
    with open(prov_file, 'w') as f:
        f.write(original_prov.replace('${PROMETHEUS_URL:-http://localhost:9090}', 'http://host.docker.internal:9090'))
    
    try:
        print("Ensuring clean environment...")
        run_compose(server_dir, ["down", "-v"])
        run_compose(display_dir, ["down", "-v"])

        print("Starting server stack...")
        run_compose(server_dir, ["up", "-d"])
        
        print("Starting display stack...")
        run_compose(display_dir, ["up", "-d"])
        
        # 3. Wait for everything to settle and scrape
        print("Waiting 20s for scrape cycle and Grafana start...")
        time.sleep(20)
        
        yield
        
    finally:
        print("Cleaning up...")
        run_compose(display_dir, ["down", "-v"])
        run_compose(server_dir, ["down", "-v"])
        with open(compose_path, 'w') as f:
            f.write(original_compose)
        with open(prov_file, 'w') as f:
            f.write(original_prov)

def test_hwexp_metrics_accessible_via_grafana_datasource():
    # Test that Grafana can actually query Prometheus for our custom metrics
    # Grafana API: /api/datasources/proxy/<id>/api/v1/query
    
    # Wait a bit more for datasource to be available
    time.sleep(5)

    # First get the datasource ID for Prometheus
    ds_resp = requests.get("http://localhost:3000/api/datasources", auth=("admin", "admin"))
    ds_resp.raise_for_status()
    datasources = ds_resp.json()
    
    prom_ds = next(ds for ds in datasources if ds["type"] == "prometheus")
    ds_id = prom_ds["id"]
    
    # Query Prometheus via Grafana proxy
    query = 'hw_device_temperature_celsius{device_class="gpu"}'
    query_url = f"http://localhost:3000/api/datasources/proxy/{ds_id}/api/v1/query?query={query}"
    
    # Retry loop for Grafana proxy
    for i in range(5):
        res = requests.get(query_url, auth=("admin", "admin"))
        if res.status_code == 200:
            break
        print(f"Grafana proxy attempt {i+1} failed ({res.status_code}): {res.text}")
        time.sleep(2)

    if res.status_code != 200:
        pytest.fail(f"Grafana proxy failed after retries: {res.status_code} {res.text}")

    data = res.json()
    
    assert data["status"] == "success"
    results = data["data"]["result"]
    assert len(results) > 0, "No metrics found in Prometheus via Grafana proxy"
    
    metric = results[0]["metric"]
    assert metric["device_class"] == "gpu"
    assert metric["component"] == "thermal"
    
    val = float(results[0]["value"][1])
    assert val == 54.234, f"Expected value 54.234, got {val}"

def test_hwexp_readyz():
    res = requests.get("http://localhost:9200/readyz")
    assert res.status_code == 200
    assert res.json()["status"] == "ready"

def test_prometheus_targets():
    res = requests.get("http://localhost:9090/api/v1/targets")
    res.raise_for_status()
    data = res.json()
    targets = data["data"]["activeTargets"]
    
    hwexp_target = next(t for t in targets if "hwexp" in t["scrapeUrl"])
    assert hwexp_target["health"] == "up"
