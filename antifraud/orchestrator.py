import json
import threading
import time
import redis
import subprocess
import os
import signal
import sys

r = redis.Redis(host="127.0.0.1", port=6379, decode_responses=True)
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))


def start_cpp_engine():
    cpp_bin = os.path.join(SCRIPT_DIR, "cpp", "fraud_engine")
    if not os.path.exists(cpp_bin):
        print("[ORCH] Building C++ engine...")
        subprocess.run(["make", "-C", os.path.join(SCRIPT_DIR, "cpp"), "build"], check=True)
    print("[ORCH] Starting C++ fraud engine...")
    return subprocess.Popen(
        [cpp_bin],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
    )


def start_python_engine():
    py_svc = os.path.join(SCRIPT_DIR, "python", "service.py")
    print("[ORCH] Starting Python fraud engine...")
    return subprocess.Popen(
        [sys.executable, py_svc],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
    )


def log_output(proc, prefix):
    for line in iter(proc.stdout.readline, ""):
        if line:
            print(f"  {prefix} {line.rstrip()}")


def orchestrator():
    print("=" * 60)
    print("  QW Pay Anti-Fraud Orchestrator")
    print("  Go -> Redis -> C++ (fast rules) + Python (deep analysis)")
    print("=" * 60)

    cpp_proc = start_cpp_engine()
    py_proc = start_python_engine()

    t1 = threading.Thread(target=log_output, args=(cpp_proc, "[CPP]"), daemon=True)
    t2 = threading.Thread(target=log_output, args=(py_proc, "[PY]"), daemon=True)
    t1.start()
    t2.start()

    def shutdown(sig, frame):
        print("\n[ORCH] Shutting down...")
        cpp_proc.terminate()
        py_proc.terminate()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    print("[ORCH] Waiting for engines to start...")
    time.sleep(2)
    print("[ORCH] All engines running. Waiting for transfer requests...")
    print("[ORCH] Queue: antifraud:queue -> C++ -> antifraud:queue:python -> Python")
    print("[ORCH] Press Ctrl+C to stop\n")

    while True:
        time.sleep(1)


if __name__ == "__main__":
    orchestrator()
