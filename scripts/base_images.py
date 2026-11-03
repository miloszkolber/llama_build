"""Resolve the base-image digests a component's Dockerfile builds from.

Base-image tags are mutable: a respin can change the image without changing
the tag. Resolving the linux/amd64 digest lets the input fingerprint notice
that refresh so a scheduled build does not keep an old base indefinitely.

Only external images are resolved. Stage references, scratch and digest-pinned
images are handled locally. Unknown registries fail loudly instead of being
silently skipped: a new registry needs a token endpoint here.
"""
import json
from pathlib import Path
import re
import sys
import time
from urllib.error import HTTPError, URLError
from urllib.parse import quote
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]
MANIFEST_ACCEPT = ", ".join([
    "application/vnd.oci.image.index.v1+json",
    "application/vnd.docker.distribution.manifest.list.v2+json",
    "application/vnd.oci.image.manifest.v1+json",
    "application/vnd.docker.distribution.manifest.v2+json",
])
PLATFORM = ("linux", "amd64")
FROM = re.compile(r"^\s*FROM\s+(?:--platform=\S+\s+)?(\S+)(?:\s+AS\s+(\S+))?\s*$",
                  re.IGNORECASE | re.MULTILINE)

# registry: (manifest endpoint, token endpoint template, repository prefix rule)
REGISTRIES = {
    "docker.io": ("registry-1.docker.io",
                  "https://auth.docker.io/token?service=registry.docker.io&scope=repository:{repo}:pull"),
    "gcr.io": ("gcr.io", "https://gcr.io/v2/token?service=gcr.io&scope=repository:{repo}:pull"),
    "ghcr.io": ("ghcr.io", "https://ghcr.io/token?service=ghcr.io&scope=repository:{repo}:pull"),
}


def parse_from_images(text):
    """Return external base images in first-use order, without duplicates."""
    aliases = {alias.lower() for _, alias in FROM.findall(text) if alias}
    images = []
    for image, _ in FROM.findall(text):
        if image.lower() in aliases or image.lower() == "scratch":
            continue
        if image.startswith("$"):
            raise ValueError("unsupported ARG-based FROM: " + image)
        if image not in images:
            images.append(image)
    return images


def split_reference(reference):
    """Split an image reference into registry, repository and tag/digest."""
    name, _, digest = reference.partition("@")
    tag = ""
    if not digest:
        last = name.rsplit("/", 1)[-1]
        if ":" in last:
            name, tag = name.rsplit(":", 1)
    if "/" in name:
        first, rest = name.split("/", 1)
        if "." in first or ":" in first or first == "localhost":
            registry, repository = first, rest
        else:
            registry, repository = "docker.io", name
    else:
        registry, repository = "docker.io", "library/" + name
    if registry == "docker.io" and "/" not in repository:
        repository = "library/" + repository
    if not tag and not digest:
        tag = "latest"
    if registry not in REGISTRIES:
        raise ValueError("unsupported base-image registry: " + registry)
    return registry, repository, digest or tag


def fetch(url, accept):
    headers = {"Accept": accept, "User-Agent": "llama_build"}
    request = Request(url, headers=headers)
    for attempt in range(2):
        try:
            with urlopen(request, timeout=30) as response:
                return response.read(), response.headers
        except (HTTPError, URLError, TimeoutError):
            if attempt == 1:
                raise
            time.sleep(2 ** attempt)
    raise RuntimeError("unreachable")


def token(registry, repository):
    _, template = REGISTRIES[registry]
    url = template.format(repo=quote(repository, safe="/"))
    body, _ = fetch(url, "application/json")
    value = json.loads(body).get("token") or json.loads(body).get("access_token")
    if not value:
        raise ValueError("registry token response has no token")
    return value


def manifest_digest(reference):
    """Return the linux/amd64 manifest digest for one base-image reference."""
    registry, repository, ref = split_reference(reference)
    if "@" in reference:
        return ref
    endpoint, _ = REGISTRIES[registry]
    auth = token(registry, repository)
    url = f"https://{endpoint}/v2/{repository}/manifests/{quote(ref, safe='')}"
    headers = {"Accept": MANIFEST_ACCEPT, "User-Agent": "llama_build",
               "Authorization": "Bearer " + auth}
    request = Request(url, headers=headers)
    for attempt in range(2):
        try:
            with urlopen(request, timeout=30) as response:
                body = response.read()
                digest = response.headers.get("Docker-Content-Digest", "")
                media = response.headers.get("Content-Type", "")
                break
        except (HTTPError, URLError, TimeoutError):
            if attempt == 1:
                raise
            time.sleep(2 ** attempt)
    document = json.loads(body)
    media = document.get("mediaType", media)
    if "index" in media or "list" in media:
        for entry in document.get("manifests", []):
            platform = entry.get("platform") or {}
            if (platform.get("os"), platform.get("architecture")) == PLATFORM:
                return entry["digest"]
        raise ValueError("no linux/amd64 manifest for " + reference)
    if not digest.startswith("sha256:"):
        raise ValueError("registry did not report a digest for " + reference)
    return digest


def resolve(component, root=ROOT):
    """Return {image reference: linux/amd64 digest} for one component."""
    dockerfile = Path(root) / "docker" / (component + ".Dockerfile")
    images = parse_from_images(dockerfile.read_text())
    return {image: manifest_digest(image) for image in images}


def main(argv):
    if len(argv) != 2 or argv[1] not in {"gate", "llama", "whisper"}:
        print("usage: base_images.py gate|llama|whisper", file=sys.stderr)
        return 2
    print(json.dumps(resolve(argv[1]), indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
