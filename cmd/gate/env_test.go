package main

import "testing"

func TestConfigFromEnv(t *testing.T) {
	env := map[string]string{"LLAMA_ARG_PORT": "9000", "WHISPER_PORT": "9001", "DOCKER_SOCKET": "/tmp/docker.sock"}
	c, err := ConfigFromEnv(func(key string) string { return env[key] })
	if err != nil {
		t.Fatal(err)
	}
	if c.DockerSocket != env["DOCKER_SOCKET"] || c.PersistentProxy.Backends[0].UpstreamPort != 9000 || c.PersistentProxy.Backends[1].UpstreamPort != 9001 {
		t.Fatal(c)
	}
	env["WHISPER_PORT"] = "70000"
	if _, err := ConfigFromEnv(func(key string) string { return env[key] }); err == nil {
		t.Fatal("invalid port accepted")
	}
}
