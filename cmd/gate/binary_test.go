package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestPersistentAlreadyCancelled(t *testing.T) {
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		c := Config{PersistentProxy: &PersistentProxyConfig{Backends: []BackendConfig{{Name: "a", ListenHost: "127.0.0.1"}, {Name: "b", ListenHost: "127.0.0.1"}}}}
		if err := NewPersistentServer(c, &fakeDocker{}).Serve(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

// Optional assembled-image acceptance: execute the binary copied from the final
// image against a mock Docker HTTP adapter and real loopback runtime sockets.
func TestGateBinaryProfiles(t *testing.T) {
	binary := os.Getenv("GATE_BINARY")
	if binary == "" {
		t.Skip("set GATE_BINARY to the final-image executable")
	}
	for _, profile := range []string{"persistent-proxy", "port-handoff"} {
		t.Run(profile, func(t *testing.T) {
			port := func() int {
				l, e := net.Listen("tcp4", "127.0.0.1:0")
				if e != nil {
					t.Fatal(e)
				}
				defer l.Close()
				return l.Addr().(*net.TCPAddr).Port
			}
			front, other, runtimePort := port(), port(), port()
			if profile == "port-handoff" {
				runtimePort = front
			}
			d := &socketRuntime{fakeDocker: fakeDocker{healthy: true}, port: runtimePort, started: make(chan struct{}, 4)}
			// Inference readiness needs HTTP; the streaming fixture returns raw bytes.
			var upstream *http.Server
			if profile == "persistent-proxy" {
				l, e := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", runtimePort))
				if e != nil {
					t.Fatal(e)
				}
				upstream = &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "runtime-ok") })}
				go upstream.Serve(l)
				defer upstream.Close()
			}
			socket := filepath.Join(t.TempDir(), "docker.sock")
			listener, e := net.Listen("unix", socket)
			if e != nil {
				t.Fatal(e)
			}
			api := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/containers/json":
					io.WriteString(w, "[]")
				case strings.HasSuffix(r.URL.Path, "/start"):
					var err error
					if profile == "port-handoff" {
						err = d.Start(r.Context(), "runtime")
					} else {
						err = d.fakeDocker.Start(r.Context(), "runtime")
					}
					if err != nil {
						http.Error(w, err.Error(), 500)
						return
					}
					w.WriteHeader(204)
				case strings.HasSuffix(r.URL.Path, "/stop"):
					if profile == "port-handoff" {
						_ = d.Stop(r.Context(), "runtime")
					} else {
						_ = d.fakeDocker.Stop(r.Context(), "runtime")
					}
					w.WriteHeader(204)
				default:
					running, _ := d.Running(r.Context(), "runtime")
					json.NewEncoder(w).Encode(map[string]any{"State": map[string]any{"Running": running, "Health": map[string]string{"Status": "healthy"}}})
				}
			})}
			go api.Serve(listener)
			defer api.Close()
			cmd := exec.Command(binary)
			cmd.Env = []string{"GATE_PROFILE=" + profile, "GATE_HOST=127.0.0.1", "DOCKER_SOCKET=" + socket, "IDLE_TIMEOUT=30", "STARTUP_TIMEOUT=2", "CONNECTION_IDLE_TIMEOUT=1", "GATE_RUNTIME=runtime", "GATE_CONTAINER_NAME=gate", fmt.Sprint("GATE_TCP_PORTS=", front), fmt.Sprint("GATE_UDP_PORTS=", front), "GATE_ACTIVITY_PREFIXES=game-", fmt.Sprint("LLAMA_PROXY_PORT=", front), fmt.Sprint("WHISPER_PROXY_PORT=", other), fmt.Sprint("LLAMA_ARG_PORT=", runtimePort), fmt.Sprint("WHISPER_PORT=", runtimePort)}
			logs, e := os.CreateTemp(t.TempDir(), "logs")
			if e != nil {
				t.Fatal(e)
			}
			defer logs.Close()
			cmd.Stdout, cmd.Stderr = logs, logs
			if e := cmd.Start(); e != nil {
				t.Fatal(e)
			}
			defer func() {
				_ = cmd.Process.Signal(syscall.SIGTERM)
				done := make(chan error, 1)
				go func() { done <- cmd.Wait() }()
				select {
				case err := <-done:
					if err != nil {
						t.Error(err)
					}
				case <-time.After(4 * time.Second):
					_ = cmd.Process.Kill()
					<-done
					t.Error("gate shutdown timed out")
				}
				_ = d.Stop(context.Background(), "runtime")
				d.wg.Wait()
			}()
			probeGate(t, front)
			d.mu.Lock()
			starts := d.starts
			d.mu.Unlock()
			if starts != 0 {
				t.Fatal("health woke runtime")
			}
			conn := dialGate(t, front)
			var err error
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(4 * time.Second))
			if profile == "port-handoff" {
				io.WriteString(conn, "ping")
				buf := make([]byte, 7)
				_, err = io.ReadFull(conn, buf)
				if err != nil || string(buf) != "ok:ping" {
					t.Fatalf("handoff %q: %v", buf, err)
				}
			} else {
				io.WriteString(conn, "GET /session HTTP/1.0\r\nHost: localhost\r\n\r\n")
				out, err := io.ReadAll(conn)
				if err != nil || !strings.Contains(string(out), "runtime-ok") {
					t.Fatalf("proxy %q: %v", out, err)
				}
			}
			d.mu.Lock()
			starts = d.starts
			d.mu.Unlock()
			if starts != 1 {
				t.Fatalf("runtime starts=%d", starts)
			}
		})
	}
}
