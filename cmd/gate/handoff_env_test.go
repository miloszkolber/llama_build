package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func environmentHandoff(t *testing.T) Config {
	t.Helper()
	env := handoffEnv()
	env["GATE_TCP_PORTS"], env["GATE_UDP_PORTS"] = "1", "1"
	c, err := ConfigFromEnv(func(key string) string { return env[key] })
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestEnvironmentHandoffBindsHostAndExcludesCustomServices(t *testing.T) {
	c := environmentHandoff(t)
	c.PortHandoff.TCPPorts, c.PortHandoff.UDPPorts = []int{0}, []int{0}
	d := &fakeDocker{names: []string{"/stream-runtime", "/stream-gate"}}
	s := NewHandoffServer(c, d)
	ls, err := s.listen()
	if err != nil {
		t.Fatal(err)
	}
	defer ls.close()
	if ls.tcp[0].Addr().(*net.TCPAddr).IP.String() != "127.0.0.1" || ls.udp[0].LocalAddr().(*net.UDPAddr).IP.String() != "127.0.0.1" {
		t.Fatal("handoff ignored GATE_HOST")
	}
	if s.games(context.Background()) {
		t.Fatal("custom gate/runtime counted as application activity")
	}
	d.names = append(d.names, "/STREAM_game_123")
	if !s.games(context.Background()) {
		t.Fatal("custom activity prefix ignored")
	}
}

// This mock Docker lifecycle owns real runtime sockets. Successful Start proves
// both gate protocols released their ports; it does not emulate Intel encoding.
type socketRuntime struct {
	fakeDocker
	port      int
	muSockets sync.Mutex
	tcp       net.Listener
	udp       *net.UDPConn
	started   chan struct{}
	wg        sync.WaitGroup
}

func (d *socketRuntime) Start(ctx context.Context, name string) error {
	d.muSockets.Lock()
	defer d.muSockets.Unlock()
	l, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", d.port))
	if err != nil {
		return err
	}
	u, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: d.port})
	if err != nil {
		l.Close()
		return err
	}
	d.tcp, d.udp = l, u
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			d.wg.Add(1)
			go func() {
				defer d.wg.Done()
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(2 * time.Second))
				buf := make([]byte, 4)
				if _, err := io.ReadFull(c, buf); err == nil {
					_, _ = c.Write(append([]byte("ok:"), buf...))
				}
			}()
		}
	}()
	if err := d.fakeDocker.Start(ctx, name); err != nil {
		return err
	}
	d.started <- struct{}{}
	return nil
}

func (d *socketRuntime) Stop(ctx context.Context, name string) error {
	d.muSockets.Lock()
	if d.tcp != nil {
		_ = d.tcp.Close()
	}
	if d.udp != nil {
		_ = d.udp.Close()
	}
	d.muSockets.Unlock()
	return d.fakeDocker.Stop(ctx, name)
}

func probeGate(t *testing.T, port int) {
	t.Helper()
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp4", fmt.Sprintf("127.0.0.1:%d", port), 100*time.Millisecond)
		if err == nil {
			_ = c.SetDeadline(time.Now().Add(200 * time.Millisecond))
			_, _ = io.WriteString(c, "GET /health HTTP/1.0\r\n\r\n")
			out, _ := io.ReadAll(c)
			_ = c.Close()
			if string(out) == "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 16\r\nConnection: close\r\n\r\n{\"status\":\"ok\"}\n" {
				return
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("gate did not answer health while owning its port")
}

func dialGate(t *testing.T, port int) net.Conn {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		conn, err := net.Dial("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return conn
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestEnvironmentHandoffColdTCPIdleReturnAndUDP(t *testing.T) {
	probe, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	probe.Close()
	c := environmentHandoff(t)
	c.PortHandoff.TCPPorts, c.PortHandoff.UDPPorts = []int{port}, []int{port}
	c.IdleTimeoutSeconds, c.StartupTimeoutSeconds, c.ConnectionIdleTimeoutSeconds = 1, 1, 1
	d := &socketRuntime{fakeDocker: fakeDocker{healthy: true}, port: port, started: make(chan struct{}, 2)}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() { result <- NewHandoffServer(c, d).Serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(3 * time.Second):
			t.Error("handoff did not cancel")
		}
		_ = d.Stop(context.Background(), c.PortHandoff.RuntimeName)
		d.wg.Wait()
	})
	probeGate(t, port)
	d.mu.Lock()
	starts := d.starts
	d.mu.Unlock()
	if starts != 0 {
		t.Fatal("health probe woke runtime")
	}
	client := dialGate(t, port)
	_ = client.SetDeadline(time.Now().Add(3 * time.Second))
	_, _ = io.WriteString(client, "ping")
	buf := make([]byte, 7)
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatal(err)
	}
	client.Close()
	if string(buf) != "ok:ping" {
		t.Fatalf("first bytes lost: %q", buf)
	}
	<-d.started
	probeGate(t, port) // Runtime stops and gate reacquires the same port.
	u, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	_, err = u.Write([]byte("trigger"))
	u.Close()
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-d.started:
	case <-time.After(2 * time.Second):
		t.Fatal("UDP did not start runtime")
	}
	d.muSockets.Lock()
	runtimeUDP := d.udp
	d.muSockets.Unlock()
	_ = runtimeUDP.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	if _, _, err := runtimeUDP.ReadFromUDP(make([]byte, 32)); err == nil {
		t.Fatal("trigger datagram was replayed")
	}
	d.mu.Lock()
	starts = d.starts
	d.mu.Unlock()
	if starts != 2 {
		t.Fatalf("starts=%d", starts)
	}
}

func TestRelayIdleActivityAndCancellation(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(fmt.Sprint(active), func(t *testing.T) {
			a, client := net.Pipe()
			b, upstream := net.Pipe()
			defer client.Close()
			defer upstream.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan struct{})
			go func() { relayConnections(ctx, a, b, 200*time.Millisecond, nil); close(done) }()
			if active {
				for i := 0; i < 4; i++ {
					time.Sleep(70 * time.Millisecond)
					go func() { _, _ = client.Write([]byte("x")) }()
					_ = upstream.SetReadDeadline(time.Now().Add(time.Second))
					buf := make([]byte, 1)
					if _, err := io.ReadFull(upstream, buf); err != nil {
						t.Fatal(err)
					}
				}
				select {
				case <-done:
					t.Fatal("activity did not refresh deadline")
				default:
				}
				cancel()
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("relay leaked after idle/cancel")
			}
		})
	}
}

func TestRelayTCPHalfClosePreservesResponse(t *testing.T) {
	pair := func() (net.Conn, *net.TCPConn) {
		l, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()
		c, err := net.DialTCP("tcp4", nil, l.Addr().(*net.TCPAddr))
		if err != nil {
			t.Fatal(err)
		}
		s, err := l.Accept()
		if err != nil {
			t.Fatal(err)
		}
		return s, c
	}
	a, client := pair()
	b, upstream := pair()
	defer client.Close()
	defer upstream.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { relayConnections(ctx, a, b, time.Second, nil); close(done) }()
	_, _ = client.Write([]byte("request"))
	_ = client.CloseWrite()
	_ = upstream.SetReadDeadline(time.Now().Add(time.Second))
	request, err := io.ReadAll(upstream)
	if err != nil || string(request) != "request" {
		t.Fatalf("%q %v", request, err)
	}
	_, _ = upstream.Write([]byte("response"))
	_ = upstream.CloseWrite()
	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	response, err := io.ReadAll(client)
	if err != nil || string(response) != "response" {
		t.Fatalf("%q %v", response, err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("half-close relay leaked")
	}
	a.Close()
	b.Close()
}
