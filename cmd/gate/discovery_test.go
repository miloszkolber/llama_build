package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func freeTCPPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func freeUDPPort(t *testing.T) int {
	t.Helper()
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	port := c.LocalAddr().(*net.UDPAddr).Port
	_ = c.Close()
	return port
}

type discoveryRuntime struct {
	fakeDocker
	port    int
	mu      sync.Mutex
	server  *http.Server
	started chan struct{}
}

func (d *discoveryRuntime) Start(ctx context.Context, name string) error {
	listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", d.port))
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/serverinfo", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "<serverinfo>cached-local-address</serverinfo>")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "runtime-answer")
	})
	d.mu.Lock()
	d.server = &http.Server{Handler: mux}
	server := d.server
	d.mu.Unlock()
	go func() { _ = server.Serve(listener) }()
	if err := d.fakeDocker.Start(ctx, name); err != nil {
		_ = server.Close()
		return err
	}
	d.started <- struct{}{}
	return nil
}

func (d *discoveryRuntime) Stop(ctx context.Context, name string) error {
	d.mu.Lock()
	server := d.server
	d.mu.Unlock()
	if server != nil {
		_ = server.Close()
	}
	return d.fakeDocker.Stop(ctx, name)
}

func requestGate(t *testing.T, port int, request string) string {
	t.Helper()
	var conn net.Conn
	var err error
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err = net.DialTimeout("tcp4", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err := io.WriteString(conn, request); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(conn)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestDiscoveryCacheAnswersWithoutWaking(t *testing.T) {
	port, udp := freeTCPPort(t), freeUDPPort(t)
	env := map[string]string{
		"GATE_PROFILE": "port-handoff", "GATE_RUNTIME": "stream-runtime",
		"GATE_CONTAINER_NAME": "stream-gate",
		"GATE_TCP_PORTS":      fmt.Sprint(port), "GATE_UDP_PORTS": fmt.Sprint(udp),
		"GATE_ACTIVITY_PREFIXES": "stream-",
		"GATE_DISCOVERY_PORT":    fmt.Sprint(port), "GATE_DISCOVERY_PATHS": "/serverinfo",
	}
	c, err := ConfigFromEnv(func(key string) string { return env[key] })
	if err != nil {
		t.Fatal(err)
	}
	c.IdleTimeoutSeconds, c.StartupTimeoutSeconds, c.ConnectionIdleTimeoutSeconds = 1, 1, 1
	d := &discoveryRuntime{fakeDocker: fakeDocker{healthy: true}, port: port, started: make(chan struct{}, 2)}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- NewHandoffServer(c, d).Serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-result:
		case <-time.After(3 * time.Second):
			t.Error("discovery handoff did not stop")
		}
	})
	if out := requestGate(t, port, "GET /serverinfo HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"); !strings.Contains(out, "cached-local-address") {
		t.Fatalf("first discovery response %q", out)
	}
	select {
	case <-d.started:
	case <-time.After(3 * time.Second):
		t.Fatal("runtime did not start")
	}
	d.mu.Lock()
	starts := d.starts
	d.mu.Unlock()
	if starts != 1 {
		t.Fatalf("starts=%d after first request", starts)
	}
	// Let the gate's idle timeout stop the runtime.
	deadline := time.Now().Add(4 * time.Second)
	for {
		d.mu.Lock()
		running := d.running
		d.mu.Unlock()
		if !running {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("runtime was not stopped after idle timeout")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if out := requestGate(t, port, "GET /serverinfo HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"); !strings.Contains(out, "cached-local-address") {
		t.Fatalf("cached discovery response %q", out)
	}
	d.mu.Lock()
	starts = d.starts
	d.mu.Unlock()
	if starts != 1 {
		t.Fatalf("cached discovery started the runtime: starts=%d", starts)
	}
	if out := requestGate(t, port, "GET /other HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n"); !strings.Contains(out, "runtime-answer") {
		t.Fatalf("non-discovery response %q", out)
	}
	d.mu.Lock()
	starts = d.starts
	d.mu.Unlock()
	if starts != 2 {
		t.Fatalf("non-discovery request did not start the runtime: starts=%d", starts)
	}
}

func TestRequestLineAndDiscoveryHost(t *testing.T) {
	method, path, host := requestLine([]byte("GET /serverinfo?x=1 HTTP/1.1\r\nHost: 192.168.0.239:47989\r\n\r\n"))
	if method != "GET" || path != "/serverinfo" || discoveryHost(host) != "192.168.0.239" {
		t.Fatalf("%q %q %q", method, path, host)
	}
	if discoveryHost("[fd00::1]:47989") != "fd00::1" || discoveryHost("Example:47989") != "example" {
		t.Fatal("host normalization failed")
	}
}

func TestCachedHeadResponseHasNoBody(t *testing.T) {
	gate, client := net.Pipe()
	defer client.Close()
	s := &HandoffServer{config: Config{PortHandoff: &PortHandoffConfig{DiscoveryPort: 1, DiscoveryPaths: []string{"/serverinfo"}}}, cache: map[string]cachedDiscovery{}}
	s.cache["127.0.0.1\x00/serverinfo"] = cachedDiscovery{body: []byte("<serverinfo/>"), contentType: "text/xml"}
	done := make(chan struct{})
	go func() {
		if !s.answerDiscovery(gate, []byte("HEAD /serverinfo HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")) {
			t.Error("cached HEAD was not answered")
		}
		close(done)
	}()
	raw, _ := io.ReadAll(client)
	<-done
	if strings.Contains(string(raw), "<serverinfo/>") || !strings.Contains(string(raw), "Content-Length: 13") {
		t.Fatalf("HEAD response %q", raw)
	}
}
