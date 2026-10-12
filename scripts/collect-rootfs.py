"""Collect matching runtime libraries, including SYCL/Intel dlopen plugins."""
import glob
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys


def collect(bindir, binary, root, backend="sycl"):
    root = Path(root)
    app = root / "app"
    (app / "lib").mkdir(parents=True)
    copied = []

    def copy(src, destination):
        dst = root / str(destination).lstrip("/")
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst, follow_symlinks=True)
        copied.append((Path(src), dst))

    copy(Path(bindir) / binary, "/app/" + binary)
    for lib in Path(bindir).glob("*.so*"):
        copy(lib, "/app/lib/" + lib.name)
    runtime_dirs = []
    if backend == "sycl":
        runtime_dirs = ["/opt/intel/oneapi/compiler/latest/lib", "/opt/intel/oneapi/mkl/latest/lib"]
        if Path("/opt/intel/oneapi/tbb/latest/lib").is_dir():
            runtime_dirs.append("/opt/intel/oneapi/tbb/latest/lib")
    for directory in runtime_dirs:
        if not Path(directory).is_dir():
            raise RuntimeError("missing oneAPI runtime directory: " + directory)
        for base, _, files in os.walk(directory, followlinks=True):
            for name in files:
                src = Path(base) / name
                if src.suffix in {".a", ".o"}:
                    continue
                copy(src, str(src))
    patterns = (["libze*.so*", "libig*.so*", "libOpenCL.so*", "libocloc.so*"]
                if backend == "sycl" else ["libvulkan.so*", "libvulkan_intel*.so*"])
    for pattern in patterns:
        matches = glob.glob("/usr/lib/x86_64-linux-gnu/" + pattern)
        if not matches:
            raise RuntimeError("missing Intel runtime libraries: " + pattern)
        for lib in matches:
            copy(lib, lib)

    env = dict(os.environ)
    env["LD_LIBRARY_PATH"] = ":".join([str(Path(bindir).resolve()), *runtime_dirs,
                                      *[d + "/intel64" for d in runtime_dirs], env.get("LD_LIBRARY_PATH", "")])
    seen = set()
    while copied:
        src, _ = copied.pop()
        resolved = src.resolve()
        if resolved in seen:
            continue
        seen.add(resolved)
        with open(src, "rb") as f:
            if f.read(4) != b"\x7fELF":
                continue
        result = subprocess.run(["ldd", str(src)], env=env, text=True, capture_output=True)
        text = result.stdout + result.stderr
        if "not found" in text:
            raise RuntimeError(f"unresolved dependencies for {src}: {text}")
        if result.returncode and "statically linked" not in text and "not a dynamic executable" not in text:
            raise RuntimeError(f"ldd failed for {src}: {text}")
        for lib in re.findall(r"(?:=>\s+)?(/[^\s()]+)", text):
            if Path(lib).is_file():
                copy(lib, lib)

    for directory in ["/etc/ssl/certs", "/etc/OpenCL/vendors"]:
        if Path(directory).exists():
            shutil.copytree(directory, root / directory.lstrip("/"), symlinks=False, dirs_exist_ok=True)
    if backend == "vulkan":
        for manifest in glob.glob("/usr/share/vulkan/icd.d/intel*.json"):
            copy(manifest, manifest)
    for copyright_file in Path("/usr/share/doc").glob("*/copyright"):
        copy(copyright_file, str(copyright_file))
    (root / "etc/passwd").write_text("runtime:x:1000:1000:runtime:/home/runtime:/nonexistent\n")
    (root / "etc/group").write_text("runtime:x:1000:\n")
    (root / "etc/nsswitch.conf").write_text("hosts: files dns\n")
    for path in ["tmp", "home/runtime", "home/runtime/.cache"]:
        directory = root / path
        directory.mkdir(parents=True, exist_ok=True)
        os.chown(directory, 1000, 1000)
    os.chmod(root / "tmp", 0o1777)
    for license_file in Path("/src").glob("LICENSE*"):
        copy(license_file, "/licenses/" + license_file.name)


if __name__ == "__main__":
    collect(*sys.argv[1:])
