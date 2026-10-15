"""Collect matching runtime libraries, including SYCL/Intel dlopen plugins."""
import glob
import fnmatch
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys


def driver_libraries(patterns, cache):
    preferred = {}
    for name, path in re.findall(r"^\s*(\S+)\s+\([^\n]*\)\s+=>\s+(/\S+)", cache, re.MULTILINE):
        preferred.setdefault(name, path)
    libraries = []
    for pattern in patterns:
        matches = [path for name, path in preferred.items() if fnmatch.fnmatchcase(name, pattern)]
        if not matches:
            raise RuntimeError("missing Intel runtime libraries: " + pattern)
        libraries.extend(matches)
    return list(dict.fromkeys(libraries))


def collect(bindir, binary, root, backend="sycl"):
    root = Path(root)
    app = root / "app"
    app.mkdir(parents=True)
    copied = []
    library_dirs = set()

    def copy(src, destination):
        dst = root / str(destination).lstrip("/")
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst, follow_symlinks=True)
        copied.append((Path(src), dst))
        if ".so" in Path(src).name:
            library_dirs.add("/" + str(dst.parent.relative_to(root)))

    copy(Path(bindir) / binary, "/app/" + binary)
    for lib in Path(bindir).glob("*.so*"):
        copy(lib, "/app/" + lib.name)
    runtime_dirs = []
    if backend == "sycl":
        runtime_dirs = ["/opt/intel/oneapi/compiler/latest/lib", "/opt/intel/oneapi/mkl/latest/lib"]
        if Path("/opt/intel/oneapi/tbb/latest/lib").is_dir():
            runtime_dirs.append("/opt/intel/oneapi/tbb/latest/lib")
    for directory in runtime_dirs:
        if not Path(directory).is_dir():
            raise RuntimeError("missing oneAPI runtime directory: " + directory)
        for base, directories, files in os.walk(directory, followlinks=False):
            for name in directories:
                src = Path(base) / name
                if src.is_symlink():
                    dst = root / str(src).lstrip("/")
                    dst.parent.mkdir(parents=True, exist_ok=True)
                    dst.symlink_to(os.readlink(src), target_is_directory=True)
            for name in files:
                src = Path(base) / name
                if src.suffix in {".a", ".o"}:
                    continue
                copy(src, str(src))
    patterns = (["libze*.so*", "libig*.so*", "libOpenCL.so*", "libocloc.so*"]
                if backend == "sycl" else ["libvulkan.so*", "libvulkan_intel*.so*"])
    # Follow the build loader's preferred SONAME, including vendor packages in
    # /usr/local/lib, instead of collecting stale distro copies by directory.
    if backend == "sycl":
        cache = subprocess.run(["ldconfig", "-p"], check=True, text=True, capture_output=True).stdout
        for lib in driver_libraries(patterns, cache):
            copy(lib, lib)
    else:
        for pattern in patterns:
            matches = glob.glob("/usr/lib/x86_64-linux-gnu/" + pattern)
            if not matches:
                raise RuntimeError("missing Intel runtime libraries: " + pattern)
            for lib in matches:
                copy(lib, lib)
    if backend == "sycl":
        # ICD drivers are loaded by filename, not linked into libOpenCL.
        for icd in Path("/etc/OpenCL/vendors").glob("*.icd"):
            for line in icd.read_text().splitlines():
                name = line.strip()
                if not name or name.startswith("#"):
                    continue
                driver = Path(name)
                if not driver.is_absolute() or not driver.is_file():
                    raise RuntimeError("missing or unsupported OpenCL ICD target: " + name)
                copy(driver, str(driver))

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
    if backend == "sycl":
        for name in ["LICENSE", "THIRD-PARTY-PROGRAMS"]:
            copy(Path("/opt/intel/oneapi/dnnl/latest/share/doc/dnnl") / name, "/licenses/onednn/" + name)
    (root / "etc/passwd").write_text("runtime:x:1000:1000:runtime:/home/runtime:/nonexistent\n")
    (root / "etc/group").write_text("runtime:x:1000:\n")
    (root / "etc/nsswitch.conf").write_text("hosts: files dns\n")
    # oneAPI dependencies can live outside the standard loader search paths.
    (root / "etc/ld.so.conf").write_text("\n".join(sorted(library_dirs)) + "\n")
    subprocess.run(["ldconfig", "-r", str(root)], check=True, capture_output=True, text=True)
    for path in ["tmp", "home/runtime", "home/runtime/.cache"]:
        directory = root / path
        directory.mkdir(parents=True, exist_ok=True)
        os.chown(directory, 1000, 1000)
    os.chmod(root / "tmp", 0o1777)
    for license_file in Path("/src").glob("LICENSE*"):
        copy(license_file, "/licenses/" + license_file.name)


if __name__ == "__main__":
    collect(*sys.argv[1:])
