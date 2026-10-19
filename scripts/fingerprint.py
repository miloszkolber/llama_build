"""Stable fingerprints over image build inputs for rebuild avoidance.

The fingerprint covers everything that can change an image without changing
its versioned tag: the resolved upstream release, this repository's
build-relevant files, and the resolved Intel driver set (llama only). The
workflow records it as an OCI label and skips scheduled rebuilds when the
published tag already carries it.

Repository documentation, the deployment example and the planning tree are
deliberately absent: editing them must not republish an unchanged image. The
path lists below decide what counts as a build input, so they have to stay in
step with the Dockerfiles; tests/test_fingerprint.py checks that every source
each Dockerfile copies is covered.
"""
import hashlib
import json
from pathlib import Path
import subprocess

LABEL = "llama-build.inputs"
ROOT = Path(__file__).resolve().parents[1]

# Files and directories that can change an image's content or how the pipeline
# builds it, per component.
BUILD_INPUTS = {
    "gate": ["docker/gate.Dockerfile", "go.mod", "cmd/gate",
             "scripts/resolve.py", ".github/workflows/build.yml"],
    "llama": ["docker/llama.Dockerfile", "scripts/collect-rootfs.py",
              "scripts/install-drivers.py", "go.mod", "cmd/probe",
              "scripts/resolve.py", ".github/workflows/build.yml"],
    "whisper": ["docker/whisper.Dockerfile", "scripts/collect-rootfs.py",
                "go.mod", "cmd/probe", "cmd/whisper",
                "scripts/resolve.py", ".github/workflows/build.yml"],
}

# Documentation that lives beside sources is copied by some Dockerfiles but
# cannot change what they build.
DOCUMENTATION_SUFFIXES = (".md",)


def drivers_digest(drivers):
    canonical = json.dumps(drivers, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def build_input_files(component):
    """Return every build-input file for a component, in a stable order."""
    files = []
    for entry in BUILD_INPUTS[component]:
        path = ROOT / entry
        if path.is_dir():
            files.extend(sorted(item for item in path.rglob("*") if item.is_file()))
        elif path.is_file():
            files.append(path)
        else:
            raise FileNotFoundError("missing build input: " + entry)
    return [item for item in files
            if not item.name.startswith("README")
            and not item.name.endswith(DOCUMENTATION_SUFFIXES)]


def repo_inputs_digest(component):
    """SHA256 over the component's build inputs, path and content together."""
    digest = hashlib.sha256()
    for path in build_input_files(component):
        digest.update(path.relative_to(ROOT).as_posix().encode("utf-8"))
        digest.update(b"\0")
        digest.update(path.read_bytes())
        digest.update(b"\0")
    return digest.hexdigest()


def fingerprint(component, tag, sha, repo_inputs, drivers=None):
    canonical = json.dumps({
        "component": component,
        "drivers": drivers_digest(drivers) if drivers is not None else None,
        "repo_inputs": repo_inputs,
        "sha": sha,
        "tag": tag,
    }, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def repo_head():
    return subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()


if __name__ == "__main__":
    import sys
    print(repo_inputs_digest(sys.argv[1]))
