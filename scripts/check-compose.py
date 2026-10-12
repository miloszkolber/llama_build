"""Check the device-release and portable Compose contract."""
import json
import sys

services = json.load(open(sys.argv[1]))["services"]
assert set(services) == {"llama", "llama-runtime", "whisper-runtime"}
assert not services["llama"].get("devices"), "gate must not own DRM devices"
assert not services["whisper-runtime"].get("command"), "Whisper arguments belong in its entrypoint"
for name, service in services.items():
    assert not any(key in service for key in ["ulimits", "cap_add", "cpuset", "read_only", "tmpfs"]), name
for name in ["llama-runtime", "whisper-runtime"]:
    assert services[name]["profiles"] == ["runtime"]
    assert services[name]["healthcheck"]["test"][1] == "/app/probe"
print("Compose topology and portability checks passed")
