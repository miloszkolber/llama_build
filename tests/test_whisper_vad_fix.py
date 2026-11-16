import importlib.util
from pathlib import Path
import subprocess
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location(
    "apply_whisper_vad_fix", Path(__file__).resolve().parents[1] / "scripts/apply_whisper_vad_fix.py"
)
fix = importlib.util.module_from_spec(spec)
spec.loader.exec_module(fix)


class WhisperVadFixTest(unittest.TestCase):
    @mock.patch.object(fix.subprocess, "run")
    def test_reviewed_revision_checks_then_applies_patch(self, run):
        fix.apply(fix.REVIEWED_SHA, "fix.patch")
        self.assertEqual(run.call_args_list, [
            mock.call(["git", "apply", "--check", "fix.patch"], check=True),
            mock.call(["git", "apply", "fix.patch"], check=True),
        ])

    @mock.patch.object(fix.subprocess, "run")
    def test_unreviewed_revision_must_already_contain_fix(self, run):
        run.return_value.returncode = 0
        fix.apply("a" * 40, "fix.patch")
        run.assert_called_once_with(
            ["git", "apply", "--reverse", "--check", "fix.patch"],
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
        )
        run.return_value.returncode = 1
        with self.assertRaisesRegex(ValueError, "review the empty-VAD fix"):
            fix.apply("b" * 40, "fix.patch")

    @mock.patch.object(fix.subprocess, "run")
    def test_patch_mismatch_does_not_apply(self, run):
        run.side_effect = subprocess.CalledProcessError(1, "git apply --check")
        with self.assertRaises(subprocess.CalledProcessError):
            fix.apply(fix.REVIEWED_SHA, "fix.patch")
        self.assertEqual(run.call_count, 1)


if __name__ == "__main__":
    unittest.main()
