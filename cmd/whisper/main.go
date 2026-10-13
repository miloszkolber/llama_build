package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"syscall"
)

func arguments(getenv func(string) string, stat func(string) (os.FileInfo, error)) ([]string, error) {
	value := func(key, fallback string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return fallback
	}
	model := getenv("WHISPER_MODEL")
	if model == "" {
		return nil, fmt.Errorf("WHISPER_MODEL is required")
	}
	port := value("WHISPER_PORT", "9670")
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return nil, fmt.Errorf("invalid WHISPER_PORT")
	}
	threads := value("WHISPER_THREADS", "4")
	n, err := strconv.Atoi(threads)
	if err != nil || n < 1 {
		return nil, fmt.Errorf("invalid WHISPER_THREADS")
	}
	args := []string{"/app/whisper-server", "--host", value("WHISPER_HOST", "127.0.0.1"),
		"--port", port, "--model", model, "--threads", threads,
		"--language", value("WHISPER_LANGUAGE", "auto"),
		"--inference-path", value("WHISPER_INFERENCE_PATH", "/v1/audio/transcriptions"),
		"--no-fallback", "--flash-attn", "--tmp-dir", "/tmp"}
	// Preserve the existing shell condition: enable VAD only with a nonempty file.
	if getenv("WHISPER_VAD_ENABLED") == "1" {
		vad := getenv("WHISPER_VAD_MODEL")
		info, err := stat(vad)
		if err == nil && info.Mode().IsRegular() && info.Size() > 0 {
			threshold := value("WHISPER_VAD_THRESHOLD", "0.5")
			v, err := strconv.ParseFloat(threshold, 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
				return nil, fmt.Errorf("invalid WHISPER_VAD_THRESHOLD")
			}
			args = append(args, "--vad", "--vad-model", vad, "--vad-threshold", threshold)
		} else {
			fmt.Fprintln(os.Stderr, "VAD requested but model missing/empty; VAD disabled")
		}
	}
	return args, nil
}

func main() {
	// Explicit command arguments bypass environment defaults, e.g. --help for smoke tests.
	args := append([]string{"/app/whisper-server"}, os.Args[1:]...)
	if len(os.Args) == 1 {
		var err error
		args, err = arguments(os.Getenv, os.Stat)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if err := syscall.Exec(args[0], args, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
