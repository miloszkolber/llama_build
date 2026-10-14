package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeDocker struct {
	mu            sync.Mutex
	running       bool
	healthy       bool
	starts, stops int
	names         []string
	listErr       error
	runningErr    error
	healthErr     error
	startErr      error
	stopErr       error
}

func (d *fakeDocker) Running(context.Context, string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running, d.runningErr
}
func (d *fakeDocker) Healthy(context.Context, string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.healthy, d.healthErr
}
func (d *fakeDocker) Start(context.Context, string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.startErr != nil {
		return d.startErr
	}
	d.starts++
	d.running = true
	return nil
}
func (d *fakeDocker) Stop(context.Context, string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopErr != nil {
		return d.stopErr
	}
	d.stops++
	d.running = false
	return nil
}
func (d *fakeDocker) ContainerNames(context.Context) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.names, d.listErr
}

type raceDocker struct {
	calls int
	stops int
}

func (d *raceDocker) Running(context.Context, string) (bool, error) {
	d.calls++
	return d.calls > 1, nil
}
func (d *raceDocker) Healthy(context.Context, string) (bool, error)    { return false, nil }
func (d *raceDocker) Start(context.Context, string) error              { return io.ErrUnexpectedEOF }
func (d *raceDocker) Stop(context.Context, string) error               { d.stops++; return nil }
func (d *raceDocker) ContainerNames(context.Context) ([]string, error) { return nil, nil }

func TestExactHealthInterception(t *testing.T) {
	for _, request := range []string{"GET /health HTTP/1.1\r\n\r\n", "HEAD /healthz HTTP/1.1\r\n\r\n"} {
		if !isHealthRequest([]byte(request)) {
			t.Fatal(request)
		}
	}
	if isHealthRequest([]byte("GET /health/ HTTP/1.1\r\n\r\n")) {
		t.Fatal("inexact health path accepted")
	}
	a, b := net.Pipe()
	go func() { writeHealthOK(a) }()
	out, _ := io.ReadAll(b)
	if !strings.Contains(string(out), "{\"status\":\"ok\"}\n") {
		t.Fatalf("unexpected health response %q", out)
	}
}
func TestPersistentStartRaceAndTimeoutCleanup(t *testing.T) {
	d := &fakeDocker{}
	c := Config{StartupTimeoutSeconds: 0, ConnectionIdleTimeoutSeconds: 1, PersistentProxy: &PersistentProxyConfig{}}
	s := NewPersistentServer(c, d)
	if s.start(context.Background(), BackendConfig{Name: "x", RuntimeName: "x", UpstreamPort: 1, HealthPath: "/"}) {
		t.Fatal("unexpected ready")
	}
	if d.stops != 1 {
		t.Fatalf("timeout cleanup stops=%d", d.stops)
	}
}

func TestPersistentStartRaceDoesNotStopExternallyStartedRuntime(t *testing.T) {
	d := &raceDocker{}
	c := Config{StartupTimeoutSeconds: 0, ConnectionIdleTimeoutSeconds: 1, PersistentProxy: &PersistentProxyConfig{}}
	if NewPersistentServer(c, d).start(context.Background(), BackendConfig{Name: "x", RuntimeName: "x", UpstreamPort: 1, HealthPath: "/"}) {
		t.Fatal("unexpected ready")
	}
	if d.calls != 2 {
		t.Fatalf("inspect calls=%d, want 2", d.calls)
	}
	if d.stops != 0 {
		t.Fatalf("externally started runtime was stopped")
	}
}

func TestPersistentStopPropagatesInspectErrors(t *testing.T) {
	b := BackendConfig{Name: "b", RuntimeName: "runtime"}
	d := &fakeDocker{running: true, runningErr: io.ErrUnexpectedEOF}
	if err := NewPersistentServer(Config{PersistentProxy: &PersistentProxyConfig{Backends: []BackendConfig{b}}}, d).stop(context.Background(), b); err == nil {
		t.Fatal("inspect error was swallowed")
	}
	d.runningErr = &dockerError{status: 404}
	if err := NewPersistentServer(Config{PersistentProxy: &PersistentProxyConfig{Backends: []BackendConfig{b}}}, d).stop(context.Background(), b); err != nil {
		t.Fatalf("404 stop = %v, want nil", err)
	}
	if d.stops != 0 {
		t.Fatalf("stop calls=%d, want 0", d.stops)
	}
}
func TestPersistentStartReinspectsAfterDockerRace(t *testing.T) {
	d := &raceDocker{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := Config{StartupTimeoutSeconds: 1, ConnectionIdleTimeoutSeconds: 1, PersistentProxy: &PersistentProxyConfig{}}
	if NewPersistentServer(c, d).start(ctx, BackendConfig{Name: "x", RuntimeName: "x", UpstreamPort: 1, HealthPath: "/"}) {
		t.Fatal("cancelled startup succeeded")
	}
	if d.calls != 2 {
		t.Fatalf("inspect calls=%d, want 2", d.calls)
	}
	if d.stops != 0 {
		t.Fatalf("externally started runtime was stopped: %d", d.stops)
	}
}
func TestHandoffActivityConservativeFailure(t *testing.T) {
	d := &fakeDocker{listErr: io.ErrUnexpectedEOF}
	s := NewHandoffServer(Config{PortHandoff: &PortHandoffConfig{RuntimeName: "wolf-runtime", ActivityPrefixes: []string{"wolf-", "wolf_", "wolf"}}}, d)
	if !s.games(context.Background()) {
		t.Fatal("list failure must count as activity")
	}
	d.listErr = nil
	d.names = []string{"/WOLF-RUNTIME", "/wolf", "/WolfSteam_123"}
	if !s.games(context.Background()) {
		t.Fatal("prefix activity not detected")
	}
	d.names = []string{"/WOLF-RUNTIME", "/wolf"}
	if s.games(context.Background()) {
		t.Fatal("runtime must be excluded")
	}
}
func TestPersistentProxyPreservesInitialBytes(t *testing.T) {
	up, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer up.Close()
	port := up.Addr().(*net.TCPAddr).Port
	go func() {
		for i := 0; i < 2; i++ {
			c, e := up.Accept()
			if e != nil {
				return
			}
			buf := make([]byte, 128)
			n, _ := c.Read(buf)
			if strings.HasPrefix(string(buf[:n]), "GET /") {
				_, _ = c.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
			} else {
				_, _ = c.Write(append([]byte("reply:"), buf[:n]...))
			}
			c.Close()
		}
	}()
	d := &fakeDocker{running: true}
	c := Config{StartupTimeoutSeconds: 1, ConnectionIdleTimeoutSeconds: 1, PersistentProxy: &PersistentProxyConfig{Backends: []BackendConfig{{Name: "b", RuntimeName: "b", UpstreamPort: port, HealthPath: "/"}}}}
	s := NewPersistentServer(c, d)
	server, client := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { s.proxy(ctx, server, c.PersistentProxy.Backends[0]); close(done) }()
	_, _ = client.Write([]byte("hello"))
	out := make([]byte, 32)
	n, _ := client.Read(out)
	_ = client.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("proxy did not finish")
	}
	if string(out[:n]) != "reply:hello" {
		t.Fatalf("got %q", out[:n])
	}
}
func TestHandoffProxyPreservesInitialBytes(t *testing.T) {
	up, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer up.Close()
	port := up.Addr().(*net.TCPAddr).Port
	go func() {
		c, _ := up.Accept()
		if c == nil {
			return
		}
		b := make([]byte, 16)
		n, _ := c.Read(b)
		_, _ = c.Write(append([]byte("ok:"), b[:n]...))
		_ = c.Close()
	}()
	s := NewHandoffServer(Config{ConnectionIdleTimeoutSeconds: 1}, &fakeDocker{})
	a, b := net.Pipe()
	done := make(chan struct{})
	go func() { s.proxy(context.Background(), a, port, []byte("tcp")); close(done) }()
	out := make([]byte, 16)
	n, _ := b.Read(out)
	_ = b.Close()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handoff proxy did not finish")
	}
	if string(out[:n]) != "ok:tcp" {
		t.Fatalf("got %q", out[:n])
	}
}

func TestHandoffStartPropagatesReadinessAndCleanupErrors(t *testing.T) {
	config := Config{StartupTimeoutSeconds: 1, PortHandoff: &PortHandoffConfig{RuntimeName: "wolf-runtime"}}
	if started, err := NewHandoffServer(config, &fakeDocker{running: true, healthErr: io.ErrUnexpectedEOF}).start(context.Background()); err == nil || started {
		t.Fatalf("readiness result=(%t,%v), want error", started, err)
	}
	config.StartupTimeoutSeconds = 0
	if started, err := NewHandoffServer(config, &fakeDocker{stopErr: io.ErrUnexpectedEOF}).start(context.Background()); err == nil || started {
		t.Fatalf("cleanup result=(%t,%v), want error", started, err)
	}
}

func TestHandoffStartRaceDoesNotStopExternallyStartedRuntime(t *testing.T) {
	d := &raceDocker{}
	config := Config{StartupTimeoutSeconds: 0, PortHandoff: &PortHandoffConfig{RuntimeName: "wolf-runtime"}}
	if started, err := NewHandoffServer(config, d).start(context.Background()); err != nil || started {
		t.Fatalf("race result=(%t,%v), want timeout without error", started, err)
	}
	if d.calls != 2 {
		t.Fatalf("inspect calls=%d, want 2", d.calls)
	}
	if d.stops != 0 {
		t.Fatalf("externally started runtime was stopped: %d", d.stops)
	}
}

func TestHandoffServePropagatesReadinessErrorAndRetriesShutdownStop(t *testing.T) {
	tcpProbe, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	tcpPort := tcpProbe.Addr().(*net.TCPAddr).Port
	_ = tcpProbe.Close()
	udpProbe, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	udpPort := udpProbe.LocalAddr().(*net.UDPAddr).Port
	_ = udpProbe.Close()
	d := &fakeDocker{healthErr: io.ErrUnexpectedEOF}
	s := NewHandoffServer(Config{IdleTimeoutSeconds: 1, StartupTimeoutSeconds: 1, PortHandoff: &PortHandoffConfig{RuntimeName: "wolf-runtime", TCPPorts: []int{tcpPort}, UDPPorts: []int{udpPort}}}, d)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- s.Serve(ctx) }()
	var client net.Conn
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		client, err = net.Dial("tcp4", fmt.Sprintf("127.0.0.1:%d", tcpPort))
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		t.Fatal(err)
	}
	_, _ = client.Write([]byte("start"))
	_ = client.Close()
	select {
	case err := <-result:
		if err == nil {
			t.Fatal("Serve swallowed readiness error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after readiness error")
	}
	if d.stops != 1 {
		t.Fatalf("shutdown stop calls=%d, want 1", d.stops)
	}
}

func TestHandoffTrafficReleasesListenerBeforeStartup(t *testing.T) {
	probe, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	_ = probe.Close()
	s := NewHandoffServer(Config{PortHandoff: &PortHandoffConfig{TCPPorts: []int{port}, UDPPorts: []int{port + 1}}}, &fakeDocker{})
	ls, err := s.listen()
	if err != nil {
		t.Fatal(err)
	}
	defer ls.close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan traffic, 1)
	go func() { events <- s.waitTraffic(ctx, ls) }()
	client, err := net.Dial("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	event := <-events
	if event.tcp == nil || event.port != port {
		t.Fatalf("event=%+v", event)
	}
	ls.close()
	_ = event.tcp.Close()
	_ = client.Close()
	runtimeListener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("port was not released before startup: %v", err)
	}
	_ = runtimeListener.Close()
}

type scriptedListener struct {
	conn    net.Conn
	release <-chan struct{}
}

func (l *scriptedListener) Accept() (net.Conn, error) { <-l.release; return l.conn, nil }
func (l *scriptedListener) Close() error              { return nil }
func (l *scriptedListener) Addr() net.Addr            { return scriptedAddr("script") }

type scriptedAddr string

func (a scriptedAddr) Network() string { return string(a) }
func (a scriptedAddr) String() string  { return string(a) }

func TestHandoffWaitTrafficClosesSimultaneousTCPLoser(t *testing.T) {
	release := make(chan struct{})
	serverA, clientA := net.Pipe()
	serverB, clientB := net.Pipe()
	defer clientA.Close()
	defer clientB.Close()
	h := &handoffListeners{tcp: []net.Listener{&scriptedListener{conn: serverA, release: release}, &scriptedListener{conn: serverB, release: release}}}
	s := NewHandoffServer(Config{PortHandoff: &PortHandoffConfig{TCPPorts: []int{1, 2}}}, &fakeDocker{})
	result := make(chan traffic, 1)
	go func() { result <- s.waitTraffic(context.Background(), h) }()
	close(release)
	winner := <-result
	if winner.tcp == nil {
		t.Fatal("missing TCP winner")
	}
	var loser net.Conn
	if winner.id == 0 {
		loser = clientB
	} else {
		loser = clientA
	}
	_ = loser.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 1)
	if n, err := loser.Read(buf); n != 0 || err == nil {
		t.Fatalf("loser remains open: n=%d err=%v", n, err)
	}
	_ = winner.tcp.Close()
}
