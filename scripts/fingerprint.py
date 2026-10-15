"""Stable fingerprints over image build inputs for rebuild avoidance.

The fingerprint covers everything that can change an image without changing
its versioned tag: the resolved upstream release, this repository's
revision, and the resolved Intel driver set (llama only). The workflow
records it as an OCI label and skips scheduled rebuilds when the published
tag already carries it.
"""
import hashlib
import json
import subprocess

LABEL = "llama-build.inputs"


def drivers_digest(drivers):
    canonical = json.dumps(drivers, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def fingerprint(component, tag, sha, repo_sha, drivers=None):
    canonical = json.dumps({
        "component": component,
        "drivers": drivers_digest(drivers) if drivers is not None else None,
        "repo": repo_sha,
        "sha": sha,
        "tag": tag,
    }, sort_keys=True, separators=(",", ":"))
    return hashlib.sha256(canonical.encode("utf-8")).hexdigest()


def repo_head():
    return subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip()
