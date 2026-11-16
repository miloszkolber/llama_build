"""Apply the upstream empty-VAD fix only to its reviewed source revision."""
import subprocess
import sys

REVIEWED_SHA = "d1be6fde11ac6e0407606b4e42fe72d34add8037"


def apply(source_sha, patch, language_patch):
    if source_sha == REVIEWED_SHA:
        subprocess.run(["git", "apply", "--check", patch], check=True)
        subprocess.run(["git", "apply", patch], check=True)
    else:
        result = subprocess.run(
            ["git", "apply", "--reverse", "--check", patch],
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
        )
        if result.returncode:
            raise ValueError("review the empty-VAD fix for this upstream revision")
        print("Upstream already contains the empty-VAD fix")
    subprocess.run(["git", "apply", "--check", language_patch], check=True)
    subprocess.run(["git", "apply", language_patch], check=True)


if __name__ == "__main__":
    if len(sys.argv) != 4:
        raise SystemExit("usage: apply_whisper_vad_fix.py SOURCE_SHA PATCH LANGUAGE_PATCH")
    apply(sys.argv[1], sys.argv[2], sys.argv[3])
