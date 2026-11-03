import importlib.util
import sys
from pathlib import Path
import unittest

ROOT = Path(__file__).parents[1]
sys.path.insert(0, str(ROOT / "scripts"))

spec = importlib.util.spec_from_file_location("base_images", ROOT / "scripts/base_images.py")
base = importlib.util.module_from_spec(spec)
spec.loader.exec_module(base)


class ParseTests(unittest.TestCase):
    def test_component_bases(self):
        expected = {
            "gate": {"golang:1.27-bookworm", "gcr.io/distroless/static-debian12:nonroot"},
            "llama": {"golang:1.27-bookworm", "intel/deep-learning-essentials:2026.2.0-devel-ubuntu24.04"},
            "whisper": {"golang:1.27-bookworm", "ubuntu:26.04"},
        }
        for component, images in expected.items():
            text = (ROOT / "docker" / (component + ".Dockerfile")).read_text()
            self.assertEqual(set(base.parse_from_images(text)), images, component)

    def test_stage_scratch_and_duplicates_are_excluded(self):
        text = """\
FROM golang:1.27-bookworm AS helper
FROM scratch
FROM helper AS build
FROM ubuntu:26.04
FROM ubuntu:26.04 AS runtime
"""
        self.assertEqual(base.parse_from_images(text), ["golang:1.27-bookworm", "ubuntu:26.04"])

    def test_platform_flag_is_accepted(self):
        self.assertEqual(
            base.parse_from_images("FROM --platform=linux/amd64 ubuntu:26.04 AS build\n"),
            ["ubuntu:26.04"])

    def test_arg_base_is_rejected(self):
        with self.assertRaises(ValueError):
            base.parse_from_images("FROM ${BASE} AS build\n")


class ReferenceTests(unittest.TestCase):
    def test_docker_hub_prefixing(self):
        self.assertEqual(base.split_reference("golang:1.27-bookworm"),
                         ("docker.io", "library/golang", "1.27-bookworm"))
        self.assertEqual(base.split_reference("intel/deep-learning-essentials:2026.2.0"),
                         ("docker.io", "intel/deep-learning-essentials", "2026.2.0"))
        self.assertEqual(base.split_reference("ubuntu:26.04"),
                         ("docker.io", "library/ubuntu", "26.04"))

    def test_other_registry(self):
        self.assertEqual(base.split_reference("gcr.io/distroless/static-debian12:nonroot"),
                         ("gcr.io", "distroless/static-debian12", "nonroot"))

    def test_digest_reference(self):
        self.assertEqual(base.split_reference("ubuntu@sha256:" + "a" * 64)[2], "sha256:" + "a" * 64)

    def test_missing_tag_defaults_to_latest(self):
        self.assertEqual(base.split_reference("ubuntu")[2], "latest")

    def test_unknown_registry_is_rejected(self):
        with self.assertRaises(ValueError):
            base.split_reference("quay.io/example/image:tag")

    def test_every_dockerfile_base_is_supported(self):
        for component in ("gate", "llama", "whisper"):
            text = (ROOT / "docker" / (component + ".Dockerfile")).read_text()
            for image in base.parse_from_images(text):
                base.split_reference(image)


class ResolveTests(unittest.TestCase):
    def test_resolve_maps_each_reference(self):
        original = base.manifest_digest
        base.manifest_digest = lambda reference: "sha256:" + "0" * 64
        try:
            resolved = base.resolve("whisper")
        finally:
            base.manifest_digest = original
        self.assertEqual(set(resolved), {"golang:1.27-bookworm", "ubuntu:26.04"})
        self.assertTrue(all(value == "sha256:" + "0" * 64 for value in resolved.values()))


if __name__ == "__main__":
    unittest.main()
