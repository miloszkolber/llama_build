import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location("policy", Path(__file__).parents[1] / "scripts/resolve.py")
policy = importlib.util.module_from_spec(spec)
spec.loader.exec_module(policy)


def release(tag, prerelease=False, draft=False):
    return dict(tag_name=tag, prerelease=prerelease, draft=draft)


class PolicyTests(unittest.TestCase):
    def test_llama_transition(self):
        releases = [release("b99", True), release("b100", True), release("v0.5.9")]
        self.assertEqual(policy.select("llama", releases), "b100")
        releases += [release("v0.6.0"), release("v0.6.1"), release("v0.7.0", True)]
        self.assertEqual(policy.select("llama", releases), "b100")
        releases += [release("v0.7.0"), release("v0.7.1")]
        self.assertEqual(policy.select("llama", releases), "v0.7.1")
        self.assertEqual(policy.select("llama", releases, "b99"), "b99")

    def test_whisper_stable_only(self):
        releases = [release("v1.9.4"), release("v1.9.5"), release("v1.10.0", True), release("v2.0.0", draft=True)]
        self.assertEqual(policy.select("whisper", releases), "v1.9.5")
        with self.assertRaises(ValueError):
            policy.select("whisper", [release("v1.9.4")])
        with self.assertRaises(ValueError):
            policy.select("whisper", releases, "master")
