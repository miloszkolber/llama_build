package main

import (
	"reflect"
	"testing"
)

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

func handoffEnv() map[string]string {
	return map[string]string{"GATE_PROFILE": "port-handoff", "GATE_RUNTIME": "stream-runtime", "GATE_CONTAINER_NAME": "stream-gate", "GATE_TCP_PORTS": "47984, 47989,48010", "GATE_UDP_PORTS": "47984,48100,48200", "GATE_ACTIVITY_PREFIXES": "stream-, stream_"}
}

func TestEnvironmentProfiles(t *testing.T) {
	c, err := ConfigFromEnv(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if c.Profile != "persistent-proxy" || c.PortHandoff != nil || c.IdleTimeoutSeconds != 300 || c.StartupTimeoutSeconds != 180 || c.ConnectionIdleTimeoutSeconds != 600 {
		t.Fatal(c)
	}
	want := []BackendConfig{{Name: "llama", RuntimeName: "llama-runtime", ListenHost: "127.0.0.1", ListenPort: 4667, UpstreamPort: 9668, HealthPath: "/health"}, {Name: "whisper", RuntimeName: "whisper-runtime", ListenHost: "127.0.0.1", ListenPort: 4670, UpstreamPort: 9670, HealthPath: "/health"}}
	if !reflect.DeepEqual(c.PersistentProxy.Backends, want) {
		t.Fatal(c.PersistentProxy)
	}
	env := handoffEnv()
	env["LLAMA_PROXY_PORT"], env["WHISPER_PORT"] = "inactive-invalid", "inactive-invalid"
	env["GATE_HOST"], env["STARTUP_TIMEOUT"] = "0.0.0.0", "90"
	c, err = ConfigFromEnv(func(k string) string { return env[k] })
	if err != nil {
		t.Fatal(err)
	}
	if c.PersistentProxy != nil || c.PortHandoff.RuntimeName != "stream-runtime" || c.PortHandoff.GateName != "stream-gate" || c.PortHandoff.ListenHost != "0.0.0.0" || c.StartupTimeoutSeconds != 90 {
		t.Fatal(c)
	}
	if !reflect.DeepEqual(c.PortHandoff.TCPPorts, []int{47984, 47989, 48010}) || !reflect.DeepEqual(c.PortHandoff.UDPPorts, []int{47984, 48100, 48200}) || !reflect.DeepEqual(c.PortHandoff.ActivityPrefixes, []string{"stream-", "stream_"}) {
		t.Fatal(c.PortHandoff)
	}
}

func TestEnvironmentRejectsInvalidSettings(t *testing.T) {
	cases := map[string][]string{
		"GATE_PROFILE": {"unknown"}, "GATE_HOST": {"bad host", "::1", "127.0.0.1:2"},
		"GATE_RUNTIME": {"", "bad/name", "stream-gate"}, "GATE_CONTAINER_NAME": {"", "bad name", "stream-runtime"},
		"GATE_TCP_PORTS": {"", "1,", "1,,2", "0", "65536", "1,1", "1,no", "999999999999999999999999"},
		"GATE_UDP_PORTS": {"", "0", "65536", "1,1"}, "GATE_ACTIVITY_PREFIXES": {"", "one,", "one, ,two"},
		"IDLE_TIMEOUT":    {"0", "-1", "bad", "9223372037", "999999999999999999999"},
		"STARTUP_TIMEOUT": {"0", "-1", "bad", "9223372037"}, "CONNECTION_IDLE_TIMEOUT": {"0", "-1", "bad", "9223372037"},
	}
	for key, values := range cases {
		for _, bad := range values {
			t.Run(key+"="+bad, func(t *testing.T) {
				env := handoffEnv()
				env[key] = bad
				if _, err := ConfigFromEnv(func(k string) string { return env[k] }); err == nil {
					t.Fatal("invalid setting accepted")
				}
			})
		}
	}
	for key, values := range map[string][]string{"LLAMA_RUNTIME": {"bad/name", "whisper-runtime"}, "WHISPER_RUNTIME": {"bad name", "llama-runtime"}, "LLAMA_PROXY_PORT": {"0", "4670", "65536"}, "WHISPER_PORT": {"0", "bad", "65536"}} {
		for _, bad := range values {
			t.Run(key+"="+bad, func(t *testing.T) {
				if _, err := ConfigFromEnv(func(k string) string {
					if k == key {
						return bad
					}
					return ""
				}); err == nil {
					t.Fatal("invalid inference setting accepted")
				}
			})
		}
	}
}
