package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"unicode"
)

// An absent profile preserves the existing two-backend inference proxy.
func ConfigFromEnv(getenv func(string) string) (Config, error) {
	value := func(key, fallback string) string {
		if v := getenv(key); v != "" {
			return v
		}
		return fallback
	}
	profile := value("GATE_PROFILE", "persistent-proxy")
	if profile != "persistent-proxy" && profile != "port-handoff" {
		return Config{}, fmt.Errorf("GATE_PROFILE must be persistent-proxy or port-handoff")
	}
	host := value("GATE_HOST", "127.0.0.1")
	ip := net.ParseIP(host)
	if ip == nil || ip.To4() == nil {
		return Config{}, fmt.Errorf("GATE_HOST must be an IPv4 address")
	}
	host = ip.To4().String()
	numbers := map[string]int{
		"IDLE_TIMEOUT": 300, "STARTUP_TIMEOUT": 180, "CONNECTION_IDLE_TIMEOUT": 600,
	}
	if profile == "persistent-proxy" {
		for key, fallback := range map[string]int{"LLAMA_PROXY_PORT": 4667, "WHISPER_PROXY_PORT": 4670, "LLAMA_ARG_PORT": 9668, "WHISPER_PORT": 9670} {
			numbers[key] = fallback
		}
	}
	for key, fallback := range numbers {
		n, err := strconv.Atoi(value(key, strconv.Itoa(fallback)))
		if err != nil {
			return Config{}, fmt.Errorf("%s must be an integer", key)
		}
		numbers[key] = n
	}
	c := Config{Profile: profile, DockerSocket: value("DOCKER_SOCKET", "/var/run/docker.sock"), IdleTimeoutSeconds: numbers["IDLE_TIMEOUT"], StartupTimeoutSeconds: numbers["STARTUP_TIMEOUT"], ConnectionIdleTimeoutSeconds: numbers["CONNECTION_IDLE_TIMEOUT"]}
	if profile == "persistent-proxy" {
		c.PersistentProxy = &PersistentProxyConfig{Backends: []BackendConfig{
			{Name: "llama", RuntimeName: value("LLAMA_RUNTIME", "llama-runtime"), ListenHost: host, ListenPort: numbers["LLAMA_PROXY_PORT"], UpstreamPort: numbers["LLAMA_ARG_PORT"], HealthPath: "/health"},
			{Name: "whisper", RuntimeName: value("WHISPER_RUNTIME", "whisper-runtime"), ListenHost: host, ListenPort: numbers["WHISPER_PROXY_PORT"], UpstreamPort: numbers["WHISPER_PORT"], HealthPath: "/health"},
		}}
	} else {
		p := &PortHandoffConfig{RuntimeName: getenv("GATE_RUNTIME"), GateName: getenv("GATE_CONTAINER_NAME"), ListenHost: host}
		if !dockerName(p.RuntimeName) || !dockerName(p.GateName) || strings.EqualFold(p.RuntimeName, p.GateName) {
			return Config{}, fmt.Errorf("GATE_RUNTIME and GATE_CONTAINER_NAME must be distinct Docker names")
		}
		var err error
		if p.TCPPorts, err = portsFromEnv("GATE_TCP_PORTS", getenv("GATE_TCP_PORTS")); err != nil {
			return Config{}, err
		}
		if p.UDPPorts, err = portsFromEnv("GATE_UDP_PORTS", getenv("GATE_UDP_PORTS")); err != nil {
			return Config{}, err
		}
		if p.ActivityPrefixes, err = listFromEnv("GATE_ACTIVITY_PREFIXES", getenv("GATE_ACTIVITY_PREFIXES")); err != nil {
			return Config{}, err
		}
		if raw := getenv("GATE_DISCOVERY_PORT"); raw != "" {
			p.DiscoveryPort, err = strconv.Atoi(raw)
			if err != nil || p.DiscoveryPort < 1 || p.DiscoveryPort > 65535 {
				return Config{}, fmt.Errorf("GATE_DISCOVERY_PORT must be a port between 1 and 65535")
			}
		}
		if raw := getenv("GATE_DISCOVERY_PATHS"); raw != "" {
			if p.DiscoveryPaths, err = listFromEnv("GATE_DISCOVERY_PATHS", raw); err != nil {
				return Config{}, err
			}
			for _, path := range p.DiscoveryPaths {
				if !validDiscoveryPath(path) {
					return Config{}, fmt.Errorf("GATE_DISCOVERY_PATHS must contain absolute HTTP paths without controls or whitespace")
				}
			}
		}
		if (p.DiscoveryPort == 0) != (len(p.DiscoveryPaths) == 0) {
			return Config{}, fmt.Errorf("GATE_DISCOVERY_PORT and GATE_DISCOVERY_PATHS must be set together")
		}
		c.PortHandoff = p
	}
	if err := validateConfig(c); err != nil {
		return Config{}, err
	}
	return c, nil
}

func listFromEnv(name, raw string) ([]string, error) {
	items := strings.Split(raw, ",")
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
		if items[i] == "" {
			return nil, fmt.Errorf("%s requires non-empty comma-separated values", name)
		}
	}
	return items, nil
}

func portsFromEnv(name, raw string) ([]int, error) {
	items, err := listFromEnv(name, raw)
	if err != nil {
		return nil, err
	}
	ports := make([]int, len(items))
	for i, item := range items {
		ports[i], err = strconv.Atoi(item)
		if err != nil {
			return nil, fmt.Errorf("%s must contain integer ports", name)
		}
	}
	if err := validPorts(ports); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return ports, nil
}

func validDiscoveryPath(path string) bool {
	if !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "?#") {
		return false
	}
	for _, r := range path {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func validateConfig(c Config) error {
	if !validTimeoutSeconds(c.IdleTimeoutSeconds) || !validTimeoutSeconds(c.StartupTimeoutSeconds) || !validTimeoutSeconds(c.ConnectionIdleTimeoutSeconds) {
		return fmt.Errorf("timeout values must be positive integer seconds without duration overflow")
	}
	if c.PortHandoff != nil && c.PortHandoff.DiscoveryPort != 0 {
		found := false
		for _, port := range c.PortHandoff.TCPPorts {
			if port == c.PortHandoff.DiscoveryPort {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("GATE_DISCOVERY_PORT must be one of GATE_TCP_PORTS")
		}
	}
	if c.PersistentProxy != nil {
		names, runtimes, ports := map[string]bool{}, map[string]bool{}, map[int]bool{}
		for _, b := range c.PersistentProxy.Backends {
			if !dockerName(b.Name) || !dockerName(b.RuntimeName) || b.ListenPort < 1 || b.ListenPort > 65535 || b.UpstreamPort < 1 || b.UpstreamPort > 65535 || !validHealthPath(b.HealthPath) {
				return fmt.Errorf("invalid persistent backend %q", b.Name)
			}
			if names[b.Name] || runtimes[b.RuntimeName] || ports[b.ListenPort] {
				return fmt.Errorf("duplicate backend name, runtime name, or listener port %q", b.Name)
			}
			names[b.Name], runtimes[b.RuntimeName], ports[b.ListenPort] = true, true, true
		}
	}
	return nil
}
