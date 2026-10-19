"""Decide whether a scheduled image build can be skipped.

Compares the fingerprint of the resolved inputs against the fingerprint
label recorded on the published versioned tag. Fails open: any missing
tag, missing label, fetch error or non-schedule trigger means build.

The fingerprint covers the resolved upstream release, the component's
build-relevant repository files and, for llama, the resolved driver
packages. Documentation-only repository changes therefore do not republish
an unchanged image.

Usage: check_inputs.py <registry-repo> <version-tag>
Environment: COMPONENT, SOURCE_TAG, SOURCE_SHA, DRIVERS_FILE
(path to drivers.json, empty when the image has none), EVENT_NAME.
Emits skip=, reason= and fingerprint= to stdout and $GITHUB_OUTPUT.
"""
import json
import os
import sys
import time
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

from fingerprint import LABEL, fingerprint, repo_inputs_digest

MANIFEST_ACCEPT = ", ".join([
    "application/vnd.oci.image.index.v1+json",
    "application/vnd.docker.distribution.manifest.list.v2+json",
    "application/vnd.oci.image.manifest.v1+json",
    "application/vnd.docker.distribution.manifest.v2+json",
])


def fetch(url, accept, token=None):
    headers = {"Accept": accept, "User-Agent": "llama_build"}
    if token:
        headers["Authorization"] = "Bearer " + token
    request = Request(url, headers=headers)
    for attempt in range(2):
        try:
            with urlopen(request, timeout=30) as response:
                return json.load(response)
        except (HTTPError, URLError, TimeoutError, json.JSONDecodeError):
            if attempt == 1:
                return None
            time.sleep(2 ** attempt)
    return None


def bearer_token(host, path):
    if host != "ghcr.io":
        return None
    body = fetch(f"https://{host}/token?service={host}&scope=repository:{path}:pull",
                 "application/json")
    return (body or {}).get("token")


def published_labels(registry_repo, tag):
    host, _, path = registry_repo.partition("/")
    if not host or not path:
        return None
    token = bearer_token(host, path)
    manifest = fetch(f"https://{host}/v2/{path}/manifests/{tag}", MANIFEST_ACCEPT, token)
    if not manifest:
        return None
    media = manifest.get("mediaType", "")
    if "index" in media or "list" in media:
        entries = manifest.get("manifests", [])
        entry = next((e for e in entries
                      if (e.get("platform") or {}).get("architecture") == "amd64"
                      and (e.get("platform") or {}).get("os") in ("linux", None)), None)
        if not entry:
            return None
        manifest = fetch(f"https://{host}/v2/{path}/manifests/{entry['digest']}",
                         "application/vnd.oci.image.manifest.v1+json", token)
        if not manifest:
            return None
    digest = ((manifest.get("config") or {}).get("digest") or "")
    if not digest:
        return None
    config = fetch(f"https://{host}/v2/{path}/blobs/{digest}", "application/json", token)
    if not config:
        return None
    return (config.get("config") or {}).get("Labels") or {}


def load_drivers(path):
    if not path:
        return None
    with open(path) as handle:
        return json.load(handle)


def decide(event, component, published, wanted):
    if component == "gate":
        return False, "gate images are untracked; always build"
    if event != "schedule":
        return False, f"explicit {event} trigger; always build"
    if published is None:
        return False, "published tag unreadable; building"
    if LABEL not in published:
        return False, "no fingerprint label on published tag; building"
    if published[LABEL] != wanted:
        return False, "inputs changed since publication; building"
    return True, "published tag already carries these inputs; skipping"


def emit(values):
    for key, value in values.items():
        print(f"{key}={value}")
    if os.getenv("GITHUB_OUTPUT"):
        with open(os.environ["GITHUB_OUTPUT"], "a") as output:
            for key, value in values.items():
                output.write(f"{key}={value}\n")


def main(argv):
    registry_repo, tag = argv[1], argv[2]
    component = os.getenv("COMPONENT", "")
    source_tag = os.getenv("SOURCE_TAG", "")
    source_sha = os.getenv("SOURCE_SHA", "")
    drivers = load_drivers(os.getenv("DRIVERS_FILE", ""))
    event = os.getenv("EVENT_NAME", "")
    wanted = fingerprint(component, source_tag, source_sha,
                         repo_inputs_digest(component), drivers)
    if component == "gate" or event != "schedule":
        published = None
    else:
        published = published_labels(registry_repo, tag)
    skip, reason = decide(event, component, published, wanted)
    emit({"skip": str(skip).lower(), "reason": reason, "fingerprint": wanted})
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
