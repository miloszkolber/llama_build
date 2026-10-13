"""Download the resolved Intel stack, checking each release asset's SHA256."""
import hashlib
import json
from pathlib import Path
import sys
from urllib.parse import urlparse
from urllib.request import urlopen

directory = Path("/tmp/debs")
directory.mkdir(exist_ok=True)
for package in json.load(open(sys.argv[1]))["packages"]:
    url = package["url"]
    parsed = urlparse(url)
    if parsed.scheme != "https" or parsed.hostname != "github.com":
        raise ValueError("unexpected release asset URL")
    filename = Path(parsed.path).name
    if not filename.endswith("_amd64.deb"):
        raise ValueError("unexpected package architecture/type")
    path = directory / filename
    digest = hashlib.sha256()
    with urlopen(url, timeout=120) as response, path.open("wb") as output:
        while chunk := response.read(1024 * 1024):
            digest.update(chunk)
            output.write(chunk)
    if digest.hexdigest() != package["sha256"]:
        path.unlink()
        raise ValueError("SHA256 mismatch: " + filename)
    print("verified", filename)
