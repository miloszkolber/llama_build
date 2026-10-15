import importlib.util
import sys
from pathlib import Path
import unittest

sys.path.insert(0, str(Path(__file__).parents[1] / "scripts"))

fingerprint_spec = importlib.util.spec_from_file_location(
    "inputs_fingerprint", Path(__file__).parents[1] / "scripts/fingerprint.py")
inputs = importlib.util.module_from_spec(fingerprint_spec)
fingerprint_spec.loader.exec_module(inputs)

check_spec = importlib.util.spec_from_file_location(
    "check_inputs", Path(__file__).parents[1] / "scripts/check_inputs.py")
check = importlib.util.module_from_spec(check_spec)
check_spec.loader.exec_module(check)


def drivers(compute="26.35.39758.10"):
    return {"compute": compute, "packages": [{"url": "https://example.invalid/a.deb", "sha256": "0" * 64}]}


class FingerprintTests(unittest.TestCase):
    def test_stable_and_hex(self):
        first = inputs.fingerprint("llama", "v0.7.0", "a" * 40, "b" * 40, drivers())
        second = inputs.fingerprint("llama", "v0.7.0", "a" * 40, "b" * 40, drivers())
        self.assertEqual(first, second)
        self.assertRegex(first, r"^[0-9a-f]{64}$")

    def test_sensitive_to_every_input(self):
        base = dict(component="llama", tag="v0.7.0", sha="a" * 40,
                    repo_sha="b" * 40, drivers=drivers())
        reference = inputs.fingerprint(**base)
        for key, value in [("component", "whisper"), ("tag", "v0.7.1"),
                           ("sha", "c" * 40), ("repo_sha", "d" * 40),
                           ("drivers", drivers("26.40.0"))]:
            altered = dict(base, **{key: value})
            self.assertNotEqual(inputs.fingerprint(**altered), reference, key)

    def test_missing_drivers_differs_from_empty(self):
        without = inputs.fingerprint("whisper", "v1.9.5", "a" * 40, "b" * 40)
        self.assertNotEqual(without, inputs.fingerprint(
            "whisper", "v1.9.5", "a" * 40, "b" * 40, drivers()))


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
