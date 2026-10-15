import importlib.util
from pathlib import Path
import unittest


spec = importlib.util.spec_from_file_location("rootfs", Path(__file__).parents[1] / "scripts/collect-rootfs.py")
rootfs = importlib.util.module_from_spec(spec)
spec.loader.exec_module(rootfs)


class DriverLibraryTests(unittest.TestCase):
    def test_preferred_vendor_igc_replaces_stale_distro_copy(self):
        cache = """4 libs found in cache `/etc/ld.so.cache'
        libigc.so.2 (libc6,x86-64) => /usr/local/lib/libigc.so.2
        libigc.so.2 (libc6,x86-64) => /lib/x86_64-linux-gnu/libigc.so.2
        libigdfcl.so.2 (libc6,x86-64) => /usr/local/lib/libigdfcl.so.2
        libigc-clang.so.23.0git (libc6,x86-64) => /usr/local/lib/libigc-clang.so.23.0git
"""
        self.assertEqual(rootfs.driver_libraries(["libig*.so*"], cache), [
            "/usr/local/lib/libigc.so.2",
            "/usr/local/lib/libigdfcl.so.2",
            "/usr/local/lib/libigc-clang.so.23.0git",
        ])

    def test_missing_required_family_fails(self):
        cache = "libze_loader.so.1 (libc6,x86-64) => /lib/x86_64-linux-gnu/libze_loader.so.1\n"
        with self.assertRaisesRegex(RuntimeError, "missing Intel runtime libraries: libig"):
            rootfs.driver_libraries(["libze*.so*", "libig*.so*"], cache)

    def test_distribution_paths_and_overlapping_patterns(self):
        cache = """libze_loader.so.1 (libc6,x86-64) => /lib/x86_64-linux-gnu/libze_loader.so.1
        libze_intel_gpu.so.1 (libc6,x86-64) => /lib/x86_64-linux-gnu/libze_intel_gpu.so.1
"""
        self.assertEqual(rootfs.driver_libraries(["libze*.so*", "libze_loader*.so*"], cache), [
            "/lib/x86_64-linux-gnu/libze_loader.so.1",
            "/lib/x86_64-linux-gnu/libze_intel_gpu.so.1",
        ])
