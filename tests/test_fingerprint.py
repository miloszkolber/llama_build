import importlib.util
import sys
from pathlib import Path
import unittest

ROOT = Path(__file__).parents[1]
sys.path.insert(0, str(ROOT / "scripts"))

fingerprint_spec = importlib.util.spec_from_file_location(
    "inputs_fingerprint", ROOT / "scripts/fingerprint.py")
inputs = importlib.util.module_from_spec(fingerprint_spec)
fingerprint_spec.loader.exec_module(inputs)

check_spec = importlib.util.spec_from_file_location(
    "check_inputs", ROOT / "scripts/check_inputs.py")
check = importlib.util.module_from_spec(check_spec)
check_spec.loader.exec_module(check)

# Artifacts the pipeline generates in the build context rather than reading
# from the repository. The resolved driver document covers their content.
GENERATED = {"llama": {".build/drivers.json"}, "whisper": set(), "gate": set()}


def drivers(compute="26.35.39758.10"):
    return {"compute": compute, "packages": [{"url": "https://example.invalid/a.deb", "sha256": "0" * 64}]}


class FingerprintTests(unittest.TestCase):
    def test_stable_and_hex(self):
        first = inputs.fingerprint("llama", "v0.7.0", "a" * 40, "b" * 64, drivers())
        second = inputs.fingerprint("llama", "v0.7.0", "a" * 40, "b" * 64, drivers())
        self.assertEqual(first, second)
        self.assertRegex(first, r"^[0-9a-f]{64}$")

    def test_sensitive_to_every_input(self):
        base = dict(component="llama", tag="v0.7.0", sha="a" * 40,
                    repo_inputs="b" * 64, drivers=drivers())
        reference = inputs.fingerprint(**base)
        for key, value in [("component", "whisper"), ("tag", "v0.7.1"),
                           ("sha", "c" * 40), ("repo_inputs", "d" * 64),
                           ("drivers", drivers("26.40.0"))]:
            altered = dict(base, **{key: value})
            self.assertNotEqual(inputs.fingerprint(**altered), reference, key)

    def test_missing_drivers_differs_from_empty(self):
        without = inputs.fingerprint("whisper", "v1.9.5", "a" * 40, "b" * 64)
        self.assertNotEqual(without, inputs.fingerprint(
            "whisper", "v1.9.5", "a" * 40, "b" * 64, drivers()))


class RepoInputsTests(unittest.TestCase):
    def test_digest_is_stable_and_per_component(self):
        first = inputs.repo_inputs_digest("llama")
        self.assertEqual(first, inputs.repo_inputs_digest("llama"))
        self.assertRegex(first, r"^[0-9a-f]{64}$")
        self.assertNotEqual(first, inputs.repo_inputs_digest("whisper"))
        self.assertNotEqual(first, inputs.repo_inputs_digest("gate"))

    def test_documentation_is_not_a_build_input(self):
        for component, entries in sorted(inputs.BUILD_INPUTS.items()):
            files = [file.relative_to(ROOT).as_posix()
                     for file in inputs.build_input_files(component)]
            self.assertNotIn("README.md", files, component)
            self.assertEqual([name for name in files if name.endswith(".md")], [], component)
            self.assertEqual([name for name in files if name.startswith("openspec/")], [], component)
            self.assertIn("go.mod", files, component)
            self.assertIn("scripts/resolve.py", files, component)
            self.assertFalse([name for name in entries if name.endswith(".md")], component)

    def test_dockerfile_copy_sources_are_covered(self):
        for component, entries in sorted(inputs.BUILD_INPUTS.items()):
            dockerfile = ROOT / "docker" / (component + ".Dockerfile")
            covered = [ROOT / entry for entry in entries]
            for line in dockerfile.read_text().splitlines():
                stripped = line.strip()
                if not stripped.upper().startswith("COPY") or "--from=" in stripped:
                    continue
                for source in stripped.split()[1:-1]:
                    if source in GENERATED[component]:
                        continue
                    path = (ROOT / source).resolve()
                    self.assertTrue(
                        any(path == entry or entry in path.parents for entry in covered),
                        f"{component}: {source} is copied by the Dockerfile but is not a build input")


class DecideTests(unittest.TestCase):
    def test_match_skips_on_schedule_only(self):
        labels = {inputs.LABEL: "fp"}
        skip, _ = check.decide("schedule", "llama", labels, "fp")
        self.assertTrue(skip)
        for event in ["push", "pull_request", "workflow_dispatch"]:
            skip, _ = check.decide(event, "llama", labels, "fp")
            self.assertFalse(skip, event)

    def test_gate_always_builds(self):
        skip, _ = check.decide("schedule", "gate", {inputs.LABEL: "fp"}, "fp")
        self.assertFalse(skip)

    def test_unknown_or_changed_inputs_build(self):
        self.assertFalse(check.decide("schedule", "llama", None, "fp")[0])
        self.assertFalse(check.decide("schedule", "llama", {}, "fp")[0])
        self.assertFalse(check.decide(
            "schedule", "llama", {inputs.LABEL: "other"}, "fp")[0])
        skip, _ = check.decide(
            "schedule", "llama", {inputs.LABEL: "fp", "other": "kept"}, "fp")
        self.assertTrue(skip)
