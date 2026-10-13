"""Resolve a versioned upstream release and its commit for image builds."""
import argparse
from datetime import datetime, timezone
import json
import os
import re
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen
from pathlib import Path
import time


def version(tag):
    match = re.fullmatch(r"v?(\d+)\.(\d+)\.(\d+)", tag)
    return tuple(map(int, match.groups())) if match else None


def select(component, releases, explicit=""):
    releases = [r for r in releases if not r.get("draft")]
    if explicit:
        if not re.fullmatch(r"(?:b\d+|v?\d+\.\d+\.\d+)", explicit):
            raise ValueError("override must be a release tag, not a branch")
        matches = [r for r in releases if r["tag_name"] == explicit]
        if not matches:
            raise ValueError("release tag not found: " + explicit)
        return matches[0]["tag_name"]
    # v0.6.0 already exists. Bootstrap on b-tags until the next feature release.
    minimum = (0, 7, 0) if component == "llama" else (1, 9, 5)
    stable = [(version(r["tag_name"]), r["tag_name"]) for r in releases
              if not r.get("prerelease") and version(r["tag_name"]) is not None
              and version(r["tag_name"]) >= minimum]
    if stable:
        return max(stable)[1]
    if component == "llama":
        nightly = [(int(r["tag_name"][1:]), r["tag_name"]) for r in releases
                   if r.get("prerelease") and re.fullmatch(r"b\d+", r["tag_name"])]
        if nightly:
            return max(nightly)[1]
    raise ValueError("no eligible upstream release")


def api(path):
    headers = {"Accept": "application/vnd.github+json", "User-Agent": "llama_build"}
    if os.getenv("GH_TOKEN"):
        headers["Authorization"] = "Bearer " + os.environ["GH_TOKEN"]
    for attempt in range(3):
        try:
            with urlopen(Request("https://api.github.com/" + path, headers=headers), timeout=30) as response:
                return json.load(response)
        except HTTPError as err:
            if err.code not in {429, 500, 502, 503, 504} or attempt == 2:
                raise
            time.sleep(2 ** attempt)


def drivers():
    compute = api("repos/intel/compute-runtime/releases/latest")
    # Follow the compiler version used by the runtime, not an unrelated latest IGC.
    match = re.search(r"intel/intel-graphics-compiler/releases/tag/([^\s)]+)", compute["body"])
    if not match:
        raise ValueError("compute-runtime release does not identify its IGC dependency")
    igc = api("repos/intel/intel-graphics-compiler/releases/tags/" + match[1])
    loader = api("repos/oneapi-src/level-zero/releases/latest")
    packages = []
    groups = [
        (compute, [r"intel-ocloc_.*_amd64.deb", r"intel-opencl-icd_.*_amd64.deb", r"libigdgmm\d+_.*_amd64.deb", r"libze-intel-gpu1_.*_amd64.deb"]),
        (igc, [r"intel-igc-core-\d+_.*_amd64.deb", r"intel-igc-opencl-\d+_.*_amd64.deb"]),
        (loader, [r"(?:libze1|level-zero)_.*u24\.04_amd64.deb", r"(?:libze-dev|level-zero-devel)_.*u24\.04_amd64.deb"]),
    ]
    for release, patterns in groups:
        for pattern in patterns:
            matches = [a for a in release["assets"] if re.fullmatch(pattern, a["name"])]
            if len(matches) != 1:
                raise ValueError(f"expected one asset matching {pattern} in {release['tag_name']}")
            asset = matches[0]
            digest = asset.get("digest", "")
            if not re.fullmatch(r"sha256:[0-9a-f]{64}", digest or ""):
                raise ValueError("asset has no SHA256 digest: " + asset["name"])
            packages.append({"url": asset["browser_download_url"], "sha256": digest[7:]})
    return {"compute": compute["tag_name"], "igc": igc["tag_name"], "level_zero": loader["tag_name"], "packages": packages}


def emit(values):
    for key, value in values.items():
        print(f"{key}={value}")
    if os.getenv("GITHUB_OUTPUT"):
        with open(os.environ["GITHUB_OUTPUT"], "a") as output:
            for key, value in values.items():
                output.write(f"{key}={value}\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("component", choices=["gate", "llama", "whisper"])
    parser.add_argument("--tag", default="")
    args = parser.parse_args()
    if args.tag and not re.fullmatch(r"(?:b\d+|v?\d+\.\d+\.\d+)", args.tag):
        parser.error("override must be a release tag")
    if args.component == "gate":
        if args.tag:
            parser.error("gate builds this repository, not an upstream release")
        emit({"sha": subprocess.check_output(["git", "rev-parse", "HEAD"], text=True).strip(), "tag": "local"})
        return
    repo = "ggml-org/" + args.component + ".cpp"
    releases = api(f"repos/{repo}/releases?per_page=100")
    # Stable releases may be older than the recent prerelease page.
    try:
        releases.append(api(f"repos/{repo}/releases/latest"))
    except HTTPError as err:
        if err.code != 404:
            raise
    if args.tag:
        releases.append(api(f"repos/{repo}/releases/tags/{args.tag}"))
    tag = select(args.component, releases, args.tag)
    refs = subprocess.check_output(["git", "ls-remote", "https://github.com/" + repo + ".git",
                                    "refs/tags/" + tag, "refs/tags/" + tag + "^{}"], text=True)
    values = dict(line.split()[::-1] for line in refs.splitlines())
    sha = values.get("refs/tags/" + tag + "^{}", values.get("refs/tags/" + tag))
    if not sha or not re.fullmatch(r"[0-9a-f]{40}", sha):
        raise ValueError("release tag did not resolve to a commit")
    if args.component == "llama":
        Path(".build").mkdir(exist_ok=True)
        Path(".build/drivers.json").write_text(json.dumps(drivers(), indent=2) + "\n")
    emit({"tag": tag, "sha": sha, "date": datetime.now(timezone.utc).date().isoformat()})


if __name__ == "__main__":
    main()
