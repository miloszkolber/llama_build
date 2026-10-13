package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArguments(t *testing.T) {
	model := "/models/model with spaces;$(false).bin"
	env := map[string]string{"WHISPER_MODEL": model}
	getenv := func(key string) string { return env[key] }
	args, err := arguments(getenv, os.Stat)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/app/whisper-server", "--host", "127.0.0.1", "--port", "9670", "--model", model,
		"--threads", "4", "--language", "auto", "--inference-path", "/v1/audio/transcriptions",
		"--no-fallback", "--flash-attn", "--tmp-dir", "/tmp"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("got %q", args)
	}
	env["WHISPER_VAD_ENABLED"] = "1"
	args, err = arguments(getenv, os.Stat)
	if err != nil || !reflect.DeepEqual(args, want) {
		t.Fatal("missing VAD should preserve base args", err)
	}
	vad := filepath.Join(t.TempDir(), "vad.bin")
	if err := os.WriteFile(vad, []byte{1}, 0600); err != nil {
		t.Fatal(err)
	}
	env["WHISPER_VAD_MODEL"] = vad
	args, err = arguments(getenv, os.Stat)
	if err != nil || !reflect.DeepEqual(args[len(want):], []string{"--vad", "--vad-model", vad, "--vad-threshold", "0.5"}) {
		t.Fatal(args, err)
	}
	for key, bad := range map[string]string{"WHISPER_PORT": "70000", "WHISPER_THREADS": "0", "WHISPER_VAD_THRESHOLD": "NaN"} {
		env[key] = bad
		if _, err := arguments(getenv, os.Stat); err == nil {
			t.Fatal("accepted", key)
		}
		delete(env, key)
	}
	delete(env, "WHISPER_MODEL")
	if _, err := arguments(getenv, os.Stat); err == nil {
		t.Fatal("missing model accepted")
	}
}
