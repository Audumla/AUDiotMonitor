from flask import Response
from prometheus_client import Counter, Gauge, generate_latest, REGISTRY
import requests
import time
import threading
import os
import re
import glob

LLAMASWAP_URL = os.environ.get("LLAMASWAP_URL", "http://localhost:41080")
POLL_INTERVAL = int(os.environ.get("LLAMASWAP_POLL_INTERVAL", "15"))
EXPORTER_PORT = int(os.environ.get("EXPORTER_PORT", "9300"))

# ── llama-swap API health ──────────────────────────────────────────────────

llamaswap_health_gauge = Gauge(
    "llamaswap_health",
    "Whether llama-swap API is reachable (1=up, 0=down)"
)

llamaswap_models_total = Gauge(
    "llamaswap_models_total",
    "Total number of models registered in llama-swap"
)

llamaswap_running_total = Gauge(
    "llamaswap_running_total",
    "Number of models currently running"
)

llamaswap_scrape_errors_total = Counter(
    "llamaswap_scrape_errors_total",
    "Total number of errors scraping llama-swap API"
)

llamaswap_backend_scrape_errors_total = Counter(
    "llamaswap_backend_scrape_errors_total",
    "Total number of errors scraping individual backend metrics",
    ["model_id"]
)

llamaswap_last_scrape_duration_seconds = Gauge(
    "llamaswap_last_scrape_duration_seconds",
    "Duration of the last llama-swap scrape in seconds"
)

gpu_busy_percent = Gauge(
    "llamaswap_gpu_busy_percent",
    "AMD GPU busy percentage from sysfs gpu_busy_percent",
    ["card"]
)

# ── llama-swap proxy-level metrics (from /running) ─────────────────────────

model_queue_length = Gauge(
    "llamaswap_model_queue_length",
    "Number of requests queued for this model",
    ["model_id"]
)

model_state = Gauge(
    "llamaswap_model_state",
    "State of the model (1=ready, 0=other/loaded)",
    ["model_id", "state"]
)

# ── llama-server backend metrics (scraped from each backend /metrics) ──────

# Throughput / current request metrics
model_requests_processing = Gauge(
    "llamaswap_requests_processing",
    "Number of requests currently processing",
    ["model_id"]
)

model_max_tokens_observed = Gauge(
    "llamaswap_max_tokens_observed",
    "Largest observed n_tokens in a single request",
    ["model_id"]
)

model_prompt_per_second = Gauge(
    "llamaswap_prompt_per_second",
    "Prompt tokens per second",
    ["model_id"]
)

model_decode_per_second = Gauge(
    "llamaswap_decode_per_second",
    "Decode tokens per second (throughput t/s)",
    ["model_id"]
)

model_busy_slots_per_decode = Gauge(
    "llamaswap_busy_slots_per_decode",
    "Average number of busy slots per llama_decode() call",
    ["model_id"]
)

model_requests_deferred = Gauge(
    "llamaswap_requests_deferred",
    "Number of requests deferred/queued",
    ["model_id"]
)

# Cumulative counters
model_prompt_tokens_total = Counter(
    "llamaswap_tokens_prompt_total",
    "Total prompt tokens processed",
    ["model_id"]
)

model_eval_tokens_total = Counter(
    "llamaswap_tokens_eval_total",
    "Total evaluated (generated) tokens processed",
    ["model_id"]
)

model_prompt_duration_total = Counter(
    "llamaswap_prompt_duration_total_seconds",
    "Total prompt evaluation duration",
    ["model_id"]
)

model_decode_duration_total = Counter(
    "llamaswap_decode_duration_total_seconds",
    "Total decode duration",
    ["model_id"]
)

model_requests_total = Counter(
    "llamaswap_requests_total",
    "Total number of requests processed",
    ["model_id"]
)

# ── Model metadata from /running + cmd parsing ─────────────────────────────

model_ttl = Gauge(
    "llamaswap_model_ttl",
    "Time to live (seconds until unload, 0=persistent)",
    ["model_id"]
)

model_description = Gauge(
    "llamaswap_model_description",
    "Model description (1=present, 0=absent)",
    ["model_id"]
)

model_name = Gauge(
    "llamaswap_model_name",
    "Model name from llama-swap config",
    ["model_id"]
)

# GPU/device info
model_device = Gauge(
    "llamaswap_model_device",
    "GPU device index (0,1,2,3)",
    ["model_id"]
)

model_device_type = Gauge(
    "llamaswap_model_device_type",
    "GPU device type (Vulkan/ROCm)",
    ["model_id"]
)

# Threading
model_threads = Gauge(
    "llamaswap_model_threads",
    "Number of CPU threads",
    ["model_id"]
)

model_threads_batch = Gauge(
    "llamaswap_model_threads_batch",
    "Number of batch CPU threads",
    ["model_id"]
)

# Sizing
model_batch_size = Gauge(
    "llamaswap_model_batch_size",
    "Logical batch size",
    ["model_id"]
)

model_ubatch_size = Gauge(
    "llamaswap_model_ubatch_size",
    "Micro batch size",
    ["model_id"]
)

model_context_length = Gauge(
    "llamaswap_model_context_length",
    "Context length (-c flag)",
    ["model_id"]
)

model_fit_target = Gauge(
    "llamaswap_model_fit_target",
    "Fit target (context slots)",
    ["model_id"]
)

model_parallel = Gauge(
    "llamaswap_model_parallel",
    "Parallel slot count",
    ["model_id"]
)

model_kv_unified = Gauge(
    "llamaswap_model_kv_unified",
    "KV cache unified mode (1=enabled)",
    ["model_id"]
)

model_cache_reuse = Gauge(
    "llamaswap_model_cache_reuse",
    "Cache reuse value",
    ["model_id"]
)

model_cache_type_k = Gauge(
    "llamaswap_model_cache_type_k",
    "KV cache type for K (encoded as numeric)",
    ["model_id"]
)

model_cache_type_v = Gauge(
    "llamaswap_model_cache_type_v",
    "KV cache type for V (encoded as numeric)",
    ["model_id"]
)

# Sampling params
model_temp = Gauge(
    "llamaswap_model_temp",
    "Temperature",
    ["model_id"]
)

model_top_p = Gauge(
    "llamaswap_model_top_p",
    "Top-p sampling parameter",
    ["model_id"]
)

model_top_k = Gauge(
    "llamaswap_model_top_k",
    "Top-k sampling parameter",
    ["model_id"]
)

model_min_p = Gauge(
    "llamaswap_model_min_p",
    "Min-p sampling parameter",
    ["model_id"]
)

model_repeat_penalty = Gauge(
    "llamaswap_model_repeat_penalty",
    "Repeat penalty",
    ["model_id"]
)

model_presence_penalty = Gauge(
    "llamaswap_model_presence_penalty",
    "Presence penalty",
    ["model_id"]
)

model_timeout = Gauge(
    "llamaswap_model_timeout",
    "Request timeout (seconds)",
    ["model_id"]
)

# Feature flags (1=enabled, 0=disabled/not-set)
model_mlock = Gauge(
    "llamaswap_model_mlock",
    "Memory lock enabled",
    ["model_id"]
)

model_cont_batching = Gauge(
    "llamaswap_model_cont_batching",
    "Continuous batching enabled",
    ["model_id"]
)

model_flash_attn = Gauge(
    "llamaswap_model_flash_attn",
    "Flash attention enabled",
    ["model_id"]
)

model_jinja = Gauge(
    "llamaswap_model_jinja",
    "Jinja template engine enabled",
    ["model_id"]
)

model_cache_prompt = Gauge(
    "llamaswap_model_cache_prompt",
    "Cache prompt enabled",
    ["model_id"]
)

model_no_context_shift = Gauge(
    "llamaswap_model_no_context_shift",
    "No context shift mode",
    ["model_id"]
)


def parse_cmd(cmd_str):
    """Parse llama-server command line to extract configuration values."""
    config = {
        "device": None,
        "device_type": None,
        "threads": None,
        "threads_batch": None,
        "batch_size": None,
        "ubatch_size": None,
        "context_length": None,
        "fit_target": None,
        "parallel": None,
        "kv_unified": 0,
        "cache_reuse": None,
        "cache_type_k": None,
        "cache_type_v": None,
        "temp": None,
        "top_p": None,
        "top_k": None,
        "min_p": None,
        "repeat_penalty": None,
        "presence_penalty": None,
        "timeout": None,
        "mlock": 0,
        "cont_batching": 0,
        "flash_attn": 0,
        "jinja": 0,
        "cache_prompt": 0,
        "no_context_shift": 0,
    }

    # Device: --device Vulkan0, --device ROCm1, etc.
    m = re.search(r'--device\s+(\w+)', cmd_str)
    if m:
        val = m.group(1)
        config["device"] = re.search(r'(\d+)', val).group(1) if re.search(r'\d+', val) else 0
        config["device_type"] = 1 if "vulkan" in val.lower() or "Vulkan" in val else 0

    # Threads
    m = re.search(r'--threads\s+(\d+)', cmd_str)
    if m:
        config["threads"] = int(m.group(1))

    m = re.search(r'--threads-batch\s+(\d+)', cmd_str)
    if m:
        config["threads_batch"] = int(m.group(1))

    # Batch sizing
    m = re.search(r'--batch-size\s+(\d+)', cmd_str)
    if m:
        config["batch_size"] = int(m.group(1))

    m = re.search(r'--ubatch-size\s+(\d+)', cmd_str)
    if m:
        config["ubatch_size"] = int(m.group(1))

    # Context length
    m = re.search(r'-c\s+(\d+)', cmd_str)
    if m:
        config["context_length"] = int(m.group(1))

    # Fit target
    m = re.search(r'--fit-target\s+(\d+)', cmd_str)
    if m:
        config["fit_target"] = int(m.group(1))

    # Parallel
    m = re.search(r'--parallel\s+(\d+)', cmd_str)
    if m:
        config["parallel"] = int(m.group(1))

    # KV unified
    if "--kv-unified" in cmd_str:
        config["kv_unified"] = 1

    # Cache reuse
    m = re.search(r'--cache-reuse\s+(\d+)', cmd_str)
    if m:
        config["cache_reuse"] = int(m.group(1))

    # Cache types
    m = re.search(r'-ctk\s+(\S+)', cmd_str)
    if m:
        config["cache_type_k"] = m.group(1)

    m = re.search(r'-ctv\s+(\S+)', cmd_str)
    if m:
        config["cache_type_v"] = m.group(1)

    # Sampling params
    m = re.search(r'--temp\s+([\d.]+)', cmd_str)
    if m:
        config["temp"] = float(m.group(1))

    m = re.search(r'--top-p\s+([\d.]+)', cmd_str)
    if m:
        config["top_p"] = float(m.group(1))

    m = re.search(r'--top-k\s+(\d+)', cmd_str)
    if m:
        config["top_k"] = int(m.group(1))

    m = re.search(r'--min-p\s+([\d.]+)', cmd_str)
    if m:
        config["min_p"] = float(m.group(1))

    m = re.search(r'--repeat-penalty\s+([\d.]+)', cmd_str)
    if m:
        config["repeat_penalty"] = float(m.group(1))

    m = re.search(r'--presence-penalty\s+([\d.]+)', cmd_str)
    if m:
        config["presence_penalty"] = float(m.group(1))

    # Timeout
    m = re.search(r'--timeout\s+(\d+)', cmd_str)
    if m:
        config["timeout"] = int(m.group(1))

    # Feature flags
    if "--mlock" in cmd_str:
        config["mlock"] = 1
    if "--cont-batching" in cmd_str:
        config["cont_batching"] = 1
    if "--flash-attn" in cmd_str:
        config["flash_attn"] = 1
    if "--jinja" in cmd_str:
        config["jinja"] = 1
    if "--cache-prompt" in cmd_str:
        config["cache_prompt"] = 1
    if "--no-context-shift" in cmd_str:
        config["no_context_shift"] = 1

    return config


def set_model_metadata(model_id, config):
    """Set Prometheus gauges from parsed model configuration."""
    if config["device"] is not None:
        model_device.labels(model_id=model_id).set(float(config["device"]))
    if config["device_type"] is not None:
        model_device_type.labels(model_id=model_id).set(float(config["device_type"]))
    if config["threads"] is not None:
        model_threads.labels(model_id=model_id).set(float(config["threads"]))
    if config["threads_batch"] is not None:
        model_threads_batch.labels(model_id=model_id).set(float(config["threads_batch"]))
    if config["batch_size"] is not None:
        model_batch_size.labels(model_id=model_id).set(float(config["batch_size"]))
    if config["ubatch_size"] is not None:
        model_ubatch_size.labels(model_id=model_id).set(float(config["ubatch_size"]))
    if config["context_length"] is not None:
        model_context_length.labels(model_id=model_id).set(float(config["context_length"]))
    if config["fit_target"] is not None:
        model_fit_target.labels(model_id=model_id).set(float(config["fit_target"]))
    if config["parallel"] is not None:
        model_parallel.labels(model_id=model_id).set(float(config["parallel"]))
    if config["kv_unified"] is not None:
        model_kv_unified.labels(model_id=model_id).set(float(config["kv_unified"]))
    if config["cache_reuse"] is not None:
        model_cache_reuse.labels(model_id=model_id).set(float(config["cache_reuse"]))
    if config["cache_type_k"] is not None:
        model_cache_type_k.labels(model_id=model_id).set(1.0)
    if config["cache_type_v"] is not None:
        model_cache_type_v.labels(model_id=model_id).set(1.0)
    if config["temp"] is not None:
        model_temp.labels(model_id=model_id).set(config["temp"])
    if config["top_p"] is not None:
        model_top_p.labels(model_id=model_id).set(config["top_p"])
    if config["top_k"] is not None:
        model_top_k.labels(model_id=model_id).set(float(config["top_k"]))
    if config["min_p"] is not None:
        model_min_p.labels(model_id=model_id).set(config["min_p"])
    if config["repeat_penalty"] is not None:
        model_repeat_penalty.labels(model_id=model_id).set(config["repeat_penalty"])
    if config["presence_penalty"] is not None:
        model_presence_penalty.labels(model_id=model_id).set(config["presence_penalty"])
    if config["timeout"] is not None:
        model_timeout.labels(model_id=model_id).set(float(config["timeout"]))
    if config["mlock"] is not None:
        model_mlock.labels(model_id=model_id).set(float(config["mlock"]))
    if config["cont_batching"] is not None:
        model_cont_batching.labels(model_id=model_id).set(float(config["cont_batching"]))
    if config["flash_attn"] is not None:
        model_flash_attn.labels(model_id=model_id).set(float(config["flash_attn"]))
    if config["jinja"] is not None:
        model_jinja.labels(model_id=model_id).set(float(config["jinja"]))
    if config["cache_prompt"] is not None:
        model_cache_prompt.labels(model_id=model_id).set(float(config["cache_prompt"]))
    if config["no_context_shift"] is not None:
        model_no_context_shift.labels(model_id=model_id).set(float(config["no_context_shift"]))


def parse_llama_server_metrics(text, model_id):
    """Parse llama-server /metrics output and set Prometheus gauges/counters."""
    for line in text.strip().split("\n"):
        line = line.strip()
        if not line or line.startswith("#"):
            continue

        # Handle histogram samples - skip
        hist_match = re.match(
            r'^(llama_(?:prompt|eval)_duration)\{quantile="([^"]+)"\}\s+([\d.eE+\-]+)$',
            line
        )
        if hist_match:
            continue

        # Handle standard metric: metric_name value  OR  metric_name{labels} value
        match = re.match(r'^([a-zA-Z_:][a-zA-Z0-9_:]*)\{(.+?)\}\s+([\d.eE+\-]+)$', line)
        if not match:
            match = re.match(r'^([a-zA-Z_:][a-zA-Z0-9_:]*)\s+([\d.eE+\-]+)$', line)
            if not match:
                continue
            name = match.group(1)
            value = float(match.group(3) if match.lastindex and match.lastindex >= 3 else match.group(2))

        # Map llama-server metrics to our gauges/counters
        if name == "llamacpp:requests_processing":
            model_requests_processing.labels(model_id=model_id).set(value)
        elif name == "llamacpp:n_tokens_max":
            model_max_tokens_observed.labels(model_id=model_id).set(value)
        elif name == "llamacpp:prompt_tokens_seconds":
            model_prompt_per_second.labels(model_id=model_id).set(value)
        elif name == "llamacpp:predicted_tokens_seconds":
            model_decode_per_second.labels(model_id=model_id).set(value)
        elif name == "llamacpp:n_busy_slots_per_decode":
            model_busy_slots_per_decode.labels(model_id=model_id).set(value)
        elif name == "llamacpp:requests_deferred":
            model_requests_deferred.labels(model_id=model_id).set(value)
        elif name == "llamacpp:prompt_tokens_total":
            model_prompt_tokens_total.labels(model_id=model_id).inc(value)
        elif name == "llamacpp:tokens_predicted_total":
            model_eval_tokens_total.labels(model_id=model_id).inc(value)
        elif name == "llamacpp:prompt_seconds_total":
            model_prompt_duration_total.labels(model_id=model_id).inc(value)
        elif name == "llamacpp:tokens_predicted_seconds_total":
            model_decode_duration_total.labels(model_id=model_id).inc(value)
        elif name == "llamacpp:n_decode_total":
            model_requests_total.labels(model_id=model_id).inc(value)


def scrape_running_models():
    """Call /running to discover active backends, then scrape each backend's /metrics."""
    try:
        resp = requests.get(f"{LLAMASWAP_URL}/running", timeout=10)
        resp.raise_for_status()
        data = resp.json()
        running = data.get("running", [])

        llamaswap_running_total.set(len(running))

        for item in running:
            model_id = item.get("model", "unknown")
            proxy = item.get("proxy", "")
            state = item.get("state", "unknown")
            cmd = item.get("cmd", "")
            ttl = item.get("ttl", 0)
            description = item.get("description", "")
            name = item.get("name", "")

            # Set proxy-level metrics
            model_queue_length.labels(model_id=model_id).set(0)
            model_state.labels(model_id=model_id, state=state).set(1.0)

            # Clear previous state for non-matching states
            for s in ["ready", "loading", "unloading"]:
                if s != state:
                    model_state.labels(model_id=model_id, state=s).set(0.0)

            # Set metadata metrics
            model_ttl.labels(model_id=model_id).set(float(ttl))
            model_description.labels(model_id=model_id).set(1.0 if description else 0.0)
            model_name.labels(model_id=model_id).set(1.0 if name else 0.0)

            # Parse cmd for configuration
            config = parse_cmd(cmd)
            set_model_metadata(model_id, config)

            # Scrape the backend's /metrics
            if proxy:
                try:
                    metrics_resp = requests.get(f"{proxy}/metrics", timeout=5)
                    if metrics_resp.status_code == 200:
                        parse_llama_server_metrics(metrics_resp.text, model_id)
                    else:
                        llamaswap_backend_scrape_errors_total.labels(model_id=model_id).inc()
                except Exception:
                    llamaswap_backend_scrape_errors_total.labels(model_id=model_id).inc()
                    # Clear metrics for this model if backend unreachable
                    for m in [model_requests_processing, model_max_tokens_observed,
                              model_prompt_per_second, model_decode_per_second,
                              model_busy_slots_per_decode, model_requests_deferred]:
                        m.labels(model_id=model_id).set(0)

    except Exception as e:
        llamaswap_scrape_errors_total.inc()
        print(f"Error scraping /running: {e}")


def scrape_model_list():
    """Call /v1/models to get registered model count."""
    try:
        resp = requests.get(f"{LLAMASWAP_URL}/v1/models", timeout=10)
        resp.raise_for_status()
        data = resp.json()
        models = data.get("data", [])
        llamaswap_models_total.set(len(models))
        llamaswap_health_gauge.set(1.0)

    except Exception as e:
        llamaswap_health_gauge.set(0.0)
        llamaswap_scrape_errors_total.inc()
        print(f"Error scraping /v1/models: {e}")


def scrape_all():
    start = time.time()
    scrape_model_list()
    scrape_running_models()
    scrape_gpu_busy()
    duration = time.time() - start
    llamaswap_last_scrape_duration_seconds.set(duration)


def scrape_gpu_busy():
    for path in glob.glob("/host/sys/class/drm/card*/device/gpu_busy_percent"):
        card = path.split("/")[-3]
        try:
            with open(path, "r", encoding="utf-8") as f:
                gpu_busy_percent.labels(card=card).set(float(f.read().strip()))
        except Exception:
            continue


def run_scrape_loop():
    scrape_all()
    while True:
        time.sleep(POLL_INTERVAL)
        scrape_all()


def handle_metrics():
    scrape_all()
    return Response(generate_latest(REGISTRY), mimetype="text/plain")


threading.Thread(target=run_scrape_loop, daemon=True).start()

if __name__ == "__main__":
    from flask import Flask
    app = Flask(__name__)
    app.add_url_rule("/metrics", "metrics", handle_metrics)
    print(f"llamaswap-exporter: llama-swap={LLAMASWAP_URL} interval={POLL_INTERVAL}s")
    print(f"llamaswap-exporter: Prometheus metrics at http://0.0.0.0:{EXPORTER_PORT}/metrics")
    app.run(host="0.0.0.0", port=EXPORTER_PORT, use_reloader=False)
