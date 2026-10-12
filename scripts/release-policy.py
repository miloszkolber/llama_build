"""Resolve a versioned upstream release and its commit for image builds."""
import argparse
from datetime import datetime
import json
import os
import re
import subprocess
from urllib.error import HTTPError
from urllib.request import Request, urlopen
from zoneinfo import ZoneInfo


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


def due(event, now):
    return event != "schedule" or now.astimezone(ZoneInfo("Europe/Warsaw")).hour == 2


def api(path):
    headers = {"Accept": "application/vnd.github+json", "User-Agent": "llama_build"}
    if os.getenv("GH_TOKEN"):
        headers["Authorization"] = "Bearer " + os.environ["GH_TOKEN"]
    with urlopen(Request("https://api.github.com/" + path, headers=headers), timeout=30) as response:
        return json.load(response)


def emit(values):
    for key, value in values.items():
        print(f"{key}={value}")
    if os.getenv("GITHUB_OUTPUT"):
        with open(os.environ["GITHUB_OUTPUT"], "a") as output:
            for key, value in values.items():
                output.write(f"{key}={value}\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("component", choices=["llama", "whisper"])
    parser.add_argument("--tag", default="")
    parser.add_argument("--event", default="workflow_dispatch")
    args = parser.parse_args()
    if not due(args.event, datetime.now(ZoneInfo("UTC"))):
        emit({"build": "false"})
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
    emit({"build": "true", "tag": tag, "sha": sha})


if __name__ == "__main__":
    main()
