package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Environment defaults mirror the existing two-backend persistent proxy.
// Explicit --config remains supported for other existing gate profiles.
func ConfigFromEnv(getenv func(string) string) (Config, error) {
	value := func(key, fallback string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return fallback
	}
	numbers := map[string]int{
		"IDLE_TIMEOUT": 300, "STARTUP_TIMEOUT": 180, "CONNECTION_IDLE_TIMEOUT": 600,
		"LLAMA_PROXY_PORT": 4667, "WHISPER_PROXY_PORT": 4670,
		"LLAMA_ARG_PORT": 9668, "WHISPER_PORT": 9670,
	}
	for key, fallback := range numbers {
		n, err := strconv.Atoi(value(key, strconv.Itoa(fallback)))
		if err != nil {
			return Config{}, fmt.Errorf("%s must be an integer", key)
		}
		numbers[key] = n
	}
	backends := PersistentProxyConfig{Backends: []BackendConfig{
		{Name: "llama", RuntimeName: value("LLAMA_RUNTIME", "llama-runtime"), ListenHost: value("GATE_HOST", "127.0.0.1"), ListenPort: numbers["LLAMA_PROXY_PORT"], UpstreamPort: numbers["LLAMA_ARG_PORT"], HealthPath: "/health"},
		{Name: "whisper", RuntimeName: value("WHISPER_RUNTIME", "whisper-runtime"), ListenHost: value("GATE_HOST", "127.0.0.1"), ListenPort: numbers["WHISPER_PROXY_PORT"], UpstreamPort: numbers["WHISPER_PORT"], HealthPath: "/health"},
	}}
	data, err := json.Marshal(map[string]any{
		"version": 1, "profile": "persistent-proxy",
		"docker_socket":        value("DOCKER_SOCKET", "/var/run/docker.sock"),
		"idle_timeout_seconds": numbers["IDLE_TIMEOUT"], "startup_timeout_seconds": numbers["STARTUP_TIMEOUT"],
		"connection_idle_timeout_seconds": numbers["CONNECTION_IDLE_TIMEOUT"], "persistent_proxy": backends,
	})
	if err != nil {
		return Config{}, err
	}
	return ParseConfig(data)
}
