package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"
)

// --- config.go ---
type Config struct {
	Profile                      string `json:"profile"`
	DockerSocket                 string `json:"docker_socket"`
	IdleTimeoutSeconds           int    `json:"idle_timeout_seconds"`
	StartupTimeoutSeconds        int    `json:"startup_timeout_seconds"`
	ConnectionIdleTimeoutSeconds int    `json:"connection_idle_timeout_seconds"`
	PersistentProxy              *PersistentProxyConfig
	PortHandoff                  *PortHandoffConfig
}

type PersistentProxyConfig struct {
	Backends []BackendConfig `json:"backends"`
}
type BackendConfig struct {
	Name         string `json:"name"`
	RuntimeName  string `json:"runtime_name"`
	ListenHost   string `json:"listen_host"`
	ListenPort   int    `json:"listen_port"`
	UpstreamPort int    `json:"upstream_port"`
	HealthPath   string `json:"health_path"`
}
type PortHandoffConfig struct {
	GateName         string
	ListenHost       string
	RuntimeName      string   `json:"runtime_name"`
	TCPPorts         []int    `json:"tcp_ports"`
	UDPPorts         []int    `json:"udp_ports"`
	ActivityPrefixes []string `json:"activity_prefixes"`
	DiscoveryPort    int
	DiscoveryPaths   []string
}

func (c Config) IdleTimeout() time.Duration { return time.Duration(c.IdleTimeoutSeconds) * time.Second }
func (c Config) StartupTimeout() time.Duration {
	return time.Duration(c.StartupTimeoutSeconds) * time.Second
}
func (c Config) ConnectionIdleTimeout() time.Duration {
	return time.Duration(c.ConnectionIdleTimeoutSeconds) * time.Second
}

func validTimeoutSeconds(seconds int) bool {
	return seconds > 0 && int64(seconds) <= int64(time.Duration(1<<63-1)/time.Second)
}

func validPorts(list []int) error {
	ports := map[int]bool{}
	for _, port := range list {
		if port < 1 || port > 65535 || ports[port] {
			return fmt.Errorf("invalid or duplicate port %d", port)
		}
		ports[port] = true
	}
	return nil
}

func dockerName(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || (first >= '0' && first <= '9')) {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-') {
			return false
		}
	}
	return true
}

func validHealthPath(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}
	for _, r := range path {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// --- common.go ---
func isHealthRequest(data []byte) bool {
	fields := strings.Fields(strings.SplitN(string(data), "\r\n", 2)[0])
	return len(fields) >= 2 && (fields[0] == "GET" || fields[0] == "HEAD") && (fields[1] == "/health" || fields[1] == "/healthz")
}
func writeHealthOK(c net.Conn) {
	_, _ = io.WriteString(c, "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: 16\r\nConnection: close\r\n\r\n{\"status\":\"ok\"}\n")
	_ = c.Close()
}
func writeUnavailable(c net.Conn) {
	_, _ = io.WriteString(c, "HTTP/1.1 503 Service Unavailable\r\nContent-Type: text/plain; charset=utf-8\r\nConnection: close\r\n\r\nRuntime is unavailable\n")
}
func inspectConnection(c net.Conn) ([]byte, bool) {
	b := make([]byte, 8192)
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _ := c.Read(b)
	_ = c.SetReadDeadline(time.Time{})
	if n == 0 {
		return nil, false
	}
	if isHealthRequest(b[:n]) {
		writeHealthOK(c)
		return nil, true
	}
	return b[:n], false
}

// requestLine identifies a plaintext HTTP request without consuming the stream.
// The gate only needs the method, path and Host to answer cached discovery.
func requestLine(data []byte) (method, path, host string) {
	head, _, _ := strings.Cut(string(data), "\r\n\r\n")
	lines := strings.Split(head, "\r\n")
	if len(lines) == 0 {
		return "", "", ""
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 2 {
		return "", "", ""
	}
	method, target := fields[0], fields[1]
	if i := strings.IndexAny(target, "?#"); i >= 0 {
		target = target[:i]
	}
	for _, line := range lines[1:] {
		name, value, ok := strings.Cut(line, ":")
		if ok && strings.EqualFold(strings.TrimSpace(name), "host") {
			host = strings.TrimSpace(value)
			break
		}
	}
	return method, target, host
}

func discoveryHost(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return strings.ToLower(strings.Trim(strings.TrimSpace(host), "[]"))
}

func (s *HandoffServer) isDiscoveryPath(path string) bool {
	for _, configured := range s.config.PortHandoff.DiscoveryPaths {
		if path == configured {
			return true
		}
	}
	return false
}

// answerDiscovery serves a cached plaintext discovery response so a client's
// periodic probe cannot wake an idle runtime. Unknown hosts fall through: the
// cached response carries the host-specific local address.
func (s *HandoffServer) answerDiscovery(c net.Conn, data []byte) bool {
	if s.config.PortHandoff.DiscoveryPort == 0 {
		return false
	}
	method, path, host := requestLine(data)
	if (method != "GET" && method != "HEAD") || !s.isDiscoveryPath(path) {
		return false
	}
	host = discoveryHost(host)
	if host == "" {
		return false
	}
	s.cacheMu.Lock()
	entry, ok := s.cache[host+"\x00"+path]
	s.cacheMu.Unlock()
	if !ok {
		return false
	}
	contentType := entry.contentType
	if contentType == "" {
		contentType = "text/xml"
	}
	header := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\nConnection: close\r\n\r\n", contentType, len(entry.body))
	if method == "HEAD" {
		_, _ = io.WriteString(c, header)
	} else {
		_, _ = c.Write(append([]byte(header), entry.body...))
	}
	_ = c.Close()
	return true
}

// refreshDiscovery fetches the configured plaintext paths from a started
// runtime. The response is cached per host so it stays valid while stopped.
func (s *HandoffServer) refreshDiscovery(ctx context.Context, host string) {
	if s.config.PortHandoff.DiscoveryPort == 0 {
		return
	}
	host = discoveryHost(host)
	if host == "" {
		return
	}
	client := &http.Client{
		Timeout:   3 * time.Second,
		Transport: &http.Transport{DisableKeepAlives: true},
	}
	for _, path := range s.config.PortHandoff.DiscoveryPaths {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d%s", s.config.PortHandoff.DiscoveryPort, path), nil)
		if err != nil {
			continue
		}
		req.Host = host
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("discovery cache %s: %v", path, err)
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 256<<10))
		_ = resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			log.Printf("discovery cache %s: status %d", path, resp.StatusCode)
			continue
		}
		s.cacheMu.Lock()
		if len(s.cache) >= 64 {
			s.cache = map[string]cachedDiscovery{}
		}
		s.cache[host+"\x00"+path] = cachedDiscovery{body: body, contentType: resp.Header.Get("Content-Type")}
		s.cacheMu.Unlock()
	}
}

func localHost(c net.Conn) string {
	if address, ok := c.LocalAddr().(*net.TCPAddr); ok {
		return address.IP.String()
	}
	return ""
}

func waitContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
func cleanupTimedOut(ctx context.Context, d Docker, runtime string) error {
	running, err := d.Running(ctx, runtime)
	if err != nil {
		return fmt.Errorf("inspect %s for timeout cleanup: %w", runtime, err)
	}
	if !running {
		return nil
	}
	if err := d.Stop(ctx, runtime); err != nil {
		return fmt.Errorf("stop timed-out %s: %w", runtime, err)
	}
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()
	for {
		runningNow, inspectErr := d.Running(ctx, runtime)
		if inspectErr != nil {
			return fmt.Errorf("inspect %s while waiting for timeout cleanup: %w", runtime, inspectErr)
		}
		if !runningNow {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			return fmt.Errorf("%s remained running after startup timeout", runtime)
		case <-tick.C:
		}
	}
}

// --- docker.go ---
type Docker interface {
	Running(context.Context, string) (bool, error)
	Healthy(context.Context, string) (bool, error)
	Start(context.Context, string) error
	Stop(context.Context, string) error
	ContainerNames(context.Context) ([]string, error)
}
type DockerClient struct{ client *http.Client }
type dockerError struct {
	status  int
	message string
}

func (e *dockerError) Error() string { return fmt.Sprintf("docker API %d: %s", e.status, e.message) }

func NewDockerClient(socket string) *DockerClient {
	t := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	return &DockerClient{client: &http.Client{Transport: t, Timeout: 30 * time.Second}}
}
func (d *DockerClient) request(ctx context.Context, method, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, "http://docker"+path, nil)
	if err != nil {
		return err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &dockerError{resp.StatusCode, strings.TrimSpace(string(b))}
	}
	if out != nil {
		return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
	}
	return nil
}

type inspectResponse struct {
	State struct {
		Running bool `json:"Running"`
		Health  *struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
}

func (d *DockerClient) inspect(ctx context.Context, name string) (inspectResponse, error) {
	var r inspectResponse
	err := d.request(ctx, http.MethodGet, "/containers/"+url.PathEscape(name)+"/json", &r)
	return r, err
}
func (d *DockerClient) Running(ctx context.Context, name string) (bool, error) {
	r, err := d.inspect(ctx, name)
	return r.State.Running, err
}
func isNotFound(err error) bool {
	e, ok := err.(*dockerError)
	return ok && e.status == http.StatusNotFound
}
func (d *DockerClient) Healthy(ctx context.Context, name string) (bool, error) {
	r, err := d.inspect(ctx, name)
	return err == nil && r.State.Running && r.State.Health != nil && r.State.Health.Status == "healthy", err
}
func (d *DockerClient) Start(ctx context.Context, name string) error {
	return d.request(ctx, http.MethodPost, "/containers/"+url.PathEscape(name)+"/start", nil)
}
func (d *DockerClient) Stop(ctx context.Context, name string) error {
	return d.request(ctx, http.MethodPost, "/containers/"+url.PathEscape(name)+"/stop?t=20", nil)
}
func (d *DockerClient) ContainerNames(ctx context.Context) ([]string, error) {
	var r []struct {
		Names []string `json:"Names"`
	}
	if err := d.request(ctx, http.MethodGet, "/containers/json", &r); err != nil {
		return nil, err
	}
	var names []string
	for _, c := range r {
		names = append(names, c.Names...)
	}
	return names, nil
}

// --- persistent.go ---
type persistentState struct {
	active   int
	last     time.Time
	starting bool
	stopping bool
	done     chan struct{}
}
type PersistentServer struct {
	config    Config
	docker    Docker
	mu        sync.Mutex
	states    map[string]*persistentState
	listeners []net.Listener
	wg        sync.WaitGroup
}

func NewPersistentServer(c Config, d Docker) *PersistentServer {
	states := make(map[string]*persistentState, len(c.PersistentProxy.Backends))
	for _, b := range c.PersistentProxy.Backends {
		states[b.Name] = &persistentState{last: time.Now()}
	}
	return &PersistentServer{config: c, docker: d, states: states}
}

func (s *PersistentServer) Serve(ctx context.Context) error {
	for _, b := range s.config.PersistentProxy.Backends {
		l, err := net.Listen("tcp4", fmt.Sprintf("%s:%d", b.ListenHost, b.ListenPort))
		if err != nil {
			s.closeListeners()
			return fmt.Errorf("listen %s: %w", b.Name, err)
		}
		s.listeners = append(s.listeners, l)
	}
	// Capture listener ownership before cancellation can clear the shared slice.
	listeners := append([]net.Listener(nil), s.listeners...)
	go func() { <-ctx.Done(); s.closeListeners() }()
	s.wg.Add(1)
	go func() { defer s.wg.Done(); s.monitor(ctx) }()
	for i, b := range s.config.PersistentProxy.Backends {
		s.wg.Add(1)
		go s.accept(ctx, listeners[i], b)
	}
	<-ctx.Done()
	s.wg.Wait()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, b := range s.config.PersistentProxy.Backends {
		if err := s.stop(shutdownCtx, b); err != nil {
			log.Printf("stop %s during shutdown: %v", b.Name, err)
		}
	}
	return nil
}
func (s *PersistentServer) closeListeners() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, l := range s.listeners {
		_ = l.Close()
	}
	s.listeners = nil
}
func (s *PersistentServer) accept(ctx context.Context, l net.Listener, b BackendConfig) {
	defer s.wg.Done()
	log.Printf("listening on port %d for %s", b.ListenPort, b.Name)
	for {
		c, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		s.wg.Add(1)
		go func() { defer s.wg.Done(); s.proxy(ctx, c, b) }()
	}
}
func (s *PersistentServer) proxy(ctx context.Context, client net.Conn, b BackendConfig) {
	defer client.Close()
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = client.Close()
		case <-done:
		}
	}()
	defer close(done)
	buffered, answered := inspectConnection(client)
	if answered || buffered == nil {
		return
	}
	if !s.acquire(ctx, b) {
		writeUnavailable(client)
		return
	}
	defer s.release(b.Name)
	upstream, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp4", fmt.Sprintf("127.0.0.1:%d", b.UpstreamPort))
	if err != nil {
		writeUnavailable(client)
		return
	}
	defer upstream.Close()
	_ = upstream.SetWriteDeadline(time.Now().Add(s.config.ConnectionIdleTimeout()))
	if _, err := upstream.Write(buffered); err != nil {
		return
	}
	s.relay(ctx, b.Name, client, upstream)
}
func (s *PersistentServer) relay(ctx context.Context, name string, a, b net.Conn) {
	relayConnections(ctx, a, b, s.config.ConnectionIdleTimeout(), func() { s.touch(name) })
}

// Both profiles share byte-activity deadlines and TCP half-close semantics.
func relayConnections(ctx context.Context, a, b net.Conn, deadline time.Duration, activity func()) {
	_ = a.SetDeadline(time.Now().Add(deadline))
	_ = b.SetDeadline(time.Now().Add(deadline))
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = a.Close()
			_ = b.Close()
		case <-done:
		}
	}()
	var wg sync.WaitGroup
	copyOne := func(src, dst net.Conn) {
		defer wg.Done()
		buf := make([]byte, 64*1024)
		for {
			n, err := src.Read(buf)
			if n > 0 {
				_ = src.SetDeadline(time.Now().Add(deadline))
				_ = dst.SetDeadline(time.Now().Add(deadline))
				if activity != nil {
					activity()
				}
				if _, writeErr := dst.Write(buf[:n]); writeErr != nil {
					_ = src.Close()
					_ = dst.Close()
					return
				}
			}
			if err != nil {
				if tcp, ok := dst.(*net.TCPConn); ok && err == io.EOF {
					_ = tcp.CloseWrite()
				} else {
					_ = src.Close()
					_ = dst.Close()
				}
				return
			}
		}
	}
	wg.Add(2)
	go copyOne(a, b)
	go copyOne(b, a)
	wg.Wait()
	close(done)
}
func (s *PersistentServer) acquire(ctx context.Context, b BackendConfig) bool {
	deadline := time.NewTimer(s.config.ConnectionIdleTimeout())
	defer deadline.Stop()
	for {
		s.mu.Lock()
		st := s.states[b.Name]
		if !st.starting && !st.stopping {
			st.starting, st.done = true, make(chan struct{})
			s.mu.Unlock()
			ok := s.start(ctx, b)
			s.mu.Lock()
			st.starting = false
			close(st.done)
			if ok {
				st.active++
				st.last = time.Now()
			}
			s.mu.Unlock()
			return ok
		}
		wait := st.done
		s.mu.Unlock()
		select {
		case <-ctx.Done():
			return false
		case <-deadline.C:
			return false
		case <-wait:
		}
	}
}
func (s *PersistentServer) release(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.states[name]
	if st.active > 0 {
		st.active--
	}
	st.last = time.Now()
}
func (s *PersistentServer) touch(name string) {
	s.mu.Lock()
	s.states[name].last = time.Now()
	s.mu.Unlock()
}
func (s *PersistentServer) start(ctx context.Context, b BackendConfig) bool {
	running, err := s.docker.Running(ctx, b.RuntimeName)
	if err != nil && !isNotFound(err) {
		log.Printf("inspect %s: %v", b.Name, err)
		return false
	}
	if isNotFound(err) {
		running = false
	}
	startedByGate := false
	if !running {
		if err := s.docker.Start(ctx, b.RuntimeName); err != nil {
			running, inspectErr := s.docker.Running(ctx, b.RuntimeName)
			if inspectErr != nil || !running {
				log.Printf("start %s: %v", b.Name, err)
				return false
			}
		} else {
			startedByGate = true
		}
	}
	deadline := time.NewTimer(s.config.StartupTimeout())
	defer deadline.Stop()
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for {
		if s.httpReady(ctx, b) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-deadline.C:
			if startedByGate {
				if err := cleanupTimedOut(ctx, s.docker, b.RuntimeName); err != nil {
					log.Printf("clean up timed-out %s: %v", b.Name, err)
				}
			}
			return false
		case <-tick.C:
		}
	}
}
func (s *PersistentServer) httpReady(ctx context.Context, b BackendConfig) bool {
	conn, err := (&net.Dialer{Timeout: 500 * time.Millisecond}).DialContext(ctx, "tcp4", fmt.Sprintf("127.0.0.1:%d", b.UpstreamPort))
	if err != nil {
		return false
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	_, _ = io.WriteString(conn, "GET "+b.HealthPath+" HTTP/1.0\r\nHost: localhost\r\n\r\n")
	buf := make([]byte, 128)
	n, _ := conn.Read(buf)
	for i := 0; i+4 <= n; i++ {
		if string(buf[i:i+4]) == " 200" {
			return true
		}
	}
	return false
}
func (s *PersistentServer) stop(ctx context.Context, b BackendConfig) error {
	running, err := s.docker.Running(ctx, b.RuntimeName)
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	return s.docker.Stop(ctx, b.RuntimeName)
}
func (s *PersistentServer) monitor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		for _, b := range s.config.PersistentProxy.Backends {
			s.mu.Lock()
			st := s.states[b.Name]
			due := st.active == 0 && !st.starting && !st.stopping && time.Since(st.last) >= s.config.IdleTimeout()
			if due {
				st.stopping = true
				st.done = make(chan struct{})
			}
			s.mu.Unlock()
			if due {
				err := s.stop(ctx, b)
				s.mu.Lock()
				st.stopping = false
				if err == nil {
					st.last = time.Now()
				}
				close(st.done)
				s.mu.Unlock()
				if err != nil {
					log.Printf("stop idle %s: %v", b.Name, err)
				}
			}
		}
	}
}

// --- handoff.go ---
type handoffListeners struct {
	tcp []net.Listener
	udp []*net.UDPConn
}

func (h *handoffListeners) close() {
	for _, l := range h.tcp {
		_ = l.Close()
	}
	for _, c := range h.udp {
		_ = c.Close()
	}
}

type traffic struct {
	tcp  net.Conn
	port int
	udp  bool
	id   int
}
type cachedDiscovery struct {
	body        []byte
	contentType string
}

type HandoffServer struct {
	config  Config
	docker  Docker
	wg      sync.WaitGroup
	cacheMu sync.Mutex
	cache   map[string]cachedDiscovery
}

func NewHandoffServer(c Config, d Docker) *HandoffServer {
	return &HandoffServer{config: c, docker: d, cache: map[string]cachedDiscovery{}}
}
func (s *HandoffServer) Serve(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer s.shutdown()
	defer func() { cancel(); s.wg.Wait() }()
	for ctx.Err() == nil {
		running, err := s.docker.Running(ctx, s.config.PortHandoff.RuntimeName)
		if err != nil {
			return fmt.Errorf("inspect runtime: %w", err)
		}
		if running {
			if err := s.monitor(ctx); err != nil {
				return err
			}
			continue
		}
		ls, err := s.listen()
		if err != nil {
			log.Printf("listen on handoff ports: %v", err)
			if !waitContext(ctx, time.Second) {
				break
			}
			continue
		}
		log.Printf("waiting on handoff ports")
		ev := s.waitTraffic(ctx, ls)
		if ev.tcp == nil && !ev.udp {
			continue
		}
		if ev.tcp != nil {
			data, answered := inspectConnection(ev.tcp)
			if answered || data == nil {
				_ = ev.tcp.Close()
				continue
			}
			if s.answerDiscovery(ev.tcp, data) {
				continue
			}
			started, err := s.start(ctx)
			if err != nil {
				_ = ev.tcp.Close()
				return err
			}
			if !started {
				_ = ev.tcp.Close()
				continue
			}
			s.refreshDiscovery(ctx, localHost(ev.tcp))
			s.wg.Add(1)
			go func() { defer s.wg.Done(); s.proxy(ctx, ev.tcp, ev.port, data) }()
		} else {
			started, err := s.start(ctx)
			if err != nil {
				return err
			}
			if !started {
				continue
			}
			if host := s.config.PortHandoff.ListenHost; host != "0.0.0.0" {
				s.refreshDiscovery(ctx, host)
			}
		}
		if err := s.monitor(ctx); err != nil {
			return err
		}
	}
	return nil
}
func (s *HandoffServer) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	running, err := s.docker.Running(ctx, s.config.PortHandoff.RuntimeName)
	if err == nil && running {
		if err := s.docker.Stop(ctx, s.config.PortHandoff.RuntimeName); err != nil {
			log.Printf("stop runtime during shutdown: %v", err)
		}
	}
}
func (s *HandoffServer) listen() (*handoffListeners, error) {
	h := &handoffListeners{}
	fail := func(e error) (*handoffListeners, error) { h.close(); return nil, e }
	for _, p := range s.config.PortHandoff.TCPPorts {
		l, e := net.Listen("tcp4", net.JoinHostPort(s.config.PortHandoff.ListenHost, fmt.Sprint(p)))
		if e != nil {
			return fail(e)
		}
		h.tcp = append(h.tcp, l)
	}
	for _, p := range s.config.PortHandoff.UDPPorts {
		c, e := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(s.config.PortHandoff.ListenHost), Port: p})
		if e != nil {
			return fail(e)
		}
		h.udp = append(h.udp, c)
	}
	return h, nil
}
func (s *HandoffServer) waitTraffic(ctx context.Context, h *handoffListeners) traffic {
	events := make(chan traffic, len(h.tcp)+len(h.udp))
	var wg sync.WaitGroup
	for i, l := range h.tcp {
		p := s.config.PortHandoff.TCPPorts[i]
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c, e := l.Accept()
			if e == nil {
				select {
				case events <- traffic{tcp: c, port: p, id: id}:
				case <-ctx.Done():
					_ = c.Close()
				}
			}
		}(i)
	}
	for i, c := range h.udp {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			b := make([]byte, 65536)
			_, _, e := c.ReadFromUDP(b)
			if e == nil {
				select {
				case events <- traffic{udp: true, id: id}:
				case <-ctx.Done():
				}
			}
		}(len(h.tcp) + i)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	var winner traffic
	gotWinner := false
	select {
	case e := <-events:
		winner, gotWinner = e, true
	case <-ctx.Done():
	case <-done:
	}
	// The runtime cannot bind until every gate listener is closed. Wait for each
	// accept/read goroutine so any concurrently accepted TCP losers are closed.
	h.close()
	wg.Wait()
	close(events)
	for event := range events {
		if event.tcp != nil && (!gotWinner || event.id != winner.id) {
			_ = event.tcp.Close()
		}
	}
	if gotWinner {
		return winner
	}
	return traffic{}
}
func (s *HandoffServer) start(ctx context.Context) (bool, error) {
	runtime := s.config.PortHandoff.RuntimeName
	running, err := s.docker.Running(ctx, runtime)
	if err != nil {
		return false, fmt.Errorf("inspect %s before start: %w", runtime, err)
	}
	startedByGate := false
	if !running {
		if err := s.docker.Start(ctx, runtime); err != nil {
			running, re := s.docker.Running(ctx, runtime)
			if re != nil {
				return false, fmt.Errorf("reinspect %s after start: %w", runtime, re)
			}
			if !running {
				return false, fmt.Errorf("start %s: %w", runtime, err)
			}
		} else {
			startedByGate = true
		}
	}
	timer := time.NewTimer(s.config.StartupTimeout())
	defer timer.Stop()
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()
	for {
		ready, err := s.ready(ctx)
		if err != nil {
			return false, fmt.Errorf("readiness for %s: %w", runtime, err)
		}
		if ready {
			return true, nil
		}
		select {
		case <-ctx.Done():
			return false, nil
		case <-timer.C:
			if startedByGate {
				if err := cleanupTimedOut(ctx, s.docker, runtime); err != nil {
					return false, err
				}
			}
			return false, nil
		case <-tick.C:
		}
	}
}
func (s *HandoffServer) ready(ctx context.Context) (bool, error) {
	healthy, err := s.docker.Healthy(ctx, s.config.PortHandoff.RuntimeName)
	if err != nil || !healthy {
		return false, err
	}
	for _, p := range s.config.PortHandoff.TCPPorts {
		c, e := (&net.Dialer{Timeout: 500 * time.Millisecond}).DialContext(ctx, "tcp4", fmt.Sprintf("127.0.0.1:%d", p))
		if e != nil {
			return false, nil
		}
		_ = c.Close()
	}
	return true, nil
}
func (s *HandoffServer) proxy(ctx context.Context, client net.Conn, port int, data []byte) {
	defer client.Close()
	up, e := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if e != nil {
		return
	}
	defer up.Close()
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = client.Close()
			_ = up.Close()
		case <-done:
		}
	}()
	defer close(done)
	_ = up.SetWriteDeadline(time.Now().Add(s.config.ConnectionIdleTimeout()))
	if _, err := up.Write(data); err != nil {
		return
	}
	relayConnections(ctx, client, up, s.config.ConnectionIdleTimeout(), nil)
}
func (s *HandoffServer) games(ctx context.Context) bool {
	names, err := s.docker.ContainerNames(ctx)
	if err != nil {
		log.Printf("list game containers: %v", err)
		return true
	}
	runtime := strings.ToLower(s.config.PortHandoff.RuntimeName)
	serviceNames := map[string]bool{runtime: true, strings.ToLower(s.config.PortHandoff.GateName): true}
	for _, suffix := range []string{"-runtime", "_runtime"} {
		if base := strings.TrimSuffix(runtime, suffix); base != "" {
			serviceNames[base] = true
		}
	}
	for _, name := range names {
		name = strings.ToLower(strings.TrimPrefix(name, "/"))
		if serviceNames[name] {
			continue
		}
		for _, prefix := range s.config.PortHandoff.ActivityPrefixes {
			if strings.HasPrefix(name, strings.ToLower(prefix)) {
				return true
			}
		}
	}
	return false
}
func (s *HandoffServer) monitor(ctx context.Context) error {
	idle := time.Now()
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		running, err := s.docker.Running(ctx, s.config.PortHandoff.RuntimeName)
		if err != nil {
			return err
		}
		if !running {
			return nil
		}
		if s.games(ctx) {
			idle = time.Now()
		}
		if time.Since(idle) >= s.config.IdleTimeout() {
			return s.docker.Stop(ctx, s.config.PortHandoff.RuntimeName)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
	}
}

// --- cmd/gate/main.go ---
func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		return
	}
	log.SetFlags(log.LstdFlags)
	health := flag.Bool("health", false, "exit successfully for container health checks")
	flag.BoolVar(health, "healthcheck", false, "alias for --health")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Printf("unexpected positional arguments")
		os.Exit(2)
	}
	if *health {
		return
	}
	c, err := ConfigFromEnv(os.Getenv)
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	docker := NewDockerClient(c.DockerSocket)
	if c.Profile == "persistent-proxy" {
		err = NewPersistentServer(c, docker).Serve(ctx)
	} else {
		err = NewHandoffServer(c, docker).Serve(ctx)
	}
	if err != nil && ctx.Err() == nil {
		log.Printf("gate failed: %v", err)
		os.Exit(1)
	}
}
