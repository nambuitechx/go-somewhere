package proxy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	// "sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Backend represents a single upstream target.
type Backend struct {
	URL          *url.URL
	Alive        int32 // 1 = alive, 0 = dead (atomic)
	// mu           sync.Mutex
	ReverseProxy *httputil.ReverseProxy
}

// IsAlive returns true if backend is marked alive.
func (b *Backend) IsAlive() bool {
	return atomic.LoadInt32(&b.Alive) == 1
}
func (b *Backend) SetAlive(alive bool) {
	if alive {
		atomic.StoreInt32(&b.Alive, 1)
	} else {
		atomic.StoreInt32(&b.Alive, 0)
	}
}

// Pool holds all backends and implements round-robin selection.
type Pool struct {
	backends []*Backend
	rrIndex  uint64 // atomic
}

func NewPool(urls []string, transport *http.Transport) (*Pool, error) {
	if len(urls) == 0 {
		return nil, errors.New("no backend URLs provided")
	}
	p := &Pool{}
	for _, s := range urls {
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid backend url %q: %w", s, err)
		}
		rp := httputil.NewSingleHostReverseProxy(u)
		// Use our custom transport for all proxied requests.
		rp.Transport = transport

		// Prevent the default Director from overwriting the raw Host header purposely.
		// We'll set headers explicitly in the proxy handler.
		backend := &Backend{
			URL:          u,
			Alive:        1, // start as alive - health checker will verify
			ReverseProxy: rp,
		}
		p.backends = append(p.backends, backend)
	}
	return p, nil
}

// NextAlive returns next alive backend using round-robin. Returns nil if none alive.
func (p *Pool) NextAlive() *Backend {
	n := uint64(len(p.backends))
	if n == 0 {
		return nil
	}
	for i := 0; i < int(n); i++ {
		idx := atomic.AddUint64(&p.rrIndex, 1)
		b := p.backends[idx%uint64(n)]
		if b.IsAlive() {
			return b
		}
	}
	return nil
}

// HealthCheck pings each backend's /healthz (or root) periodically.
// If healthPath is empty, "/" is used.
func (p *Pool) HealthCheck(ctx context.Context, interval time.Duration, timeout time.Duration, healthPath string) {
	if healthPath == "" {
		healthPath = "/healthz"
	}
	client := &http.Client{Timeout: timeout}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, b := range p.backends {
				go func(be *Backend) {
					u := *be.URL // copy
					u.Path = healthPath
					req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
					if err != nil {
						log.Printf("[health] create request err=%v for %s", err, u.String())
						be.SetAlive(false)
						return
					}
					resp, err := client.Do(req)
					if err != nil {
						log.Printf("[health] %s unhealthy: %v", be.URL, err)
						be.SetAlive(false)
						return
					}
					resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						if !be.IsAlive() {
							log.Printf("[health] %s became alive (status=%d)", be.URL, resp.StatusCode)
						}
						be.SetAlive(true)
						return
					}
					log.Printf("[health] %s returned status=%d", be.URL, resp.StatusCode)
					be.SetAlive(false)
				}(b)
			}
		}
	}
}

// removeHopByHopHeaders removes headers that should not be forwarded.
func removeHopByHopHeaders(h http.Header) {
	// As per RFC 7230 section 6.1: Connection-specific headers should be removed.
	hopHeaders := []string{
		"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization",
		"TE", "Trailers", "Transfer-Encoding", "Upgrade",
	}
	for _, hh := range hopHeaders {
		h.Del(hh)
	}
	// Also remove entries listed in Connection header
	if conn := h.Get("Connection"); conn != "" {
		for _, token := range strings.Split(conn, ",") {
			if t := strings.TrimSpace(token); t != "" {
				h.Del(t)
			}
		}
	}
}

// ProxyHandler returns an http.Handler that proxies requests to selected backend.
func ProxyHandler(pool *Pool, reqTimeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		backend := pool.NextAlive()
		if backend == nil {
			http.Error(w, "no upstream available", http.StatusServiceUnavailable)
			log.Printf("[proxy] no upstream available for %s %s", r.Method, r.URL.Path)
			return
		}

		// Setup request context with timeout to avoid leaking handlers
		ctx, cancel := context.WithTimeout(r.Context(), reqTimeout)
		defer cancel()
		r = r.WithContext(ctx)

		// Prepare director-like modifications:
		// - set URL scheme/host/path based on backend
		// - preserve original request path (no default Director use)
		// - forward X-Forwarded-For / Proto / Host
		proxy := backend.ReverseProxy

		// Custom Director: don't overwrite URL.Path with backend.Path if backend has prefix.
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			// call original director to setup basics (URL.Scheme, Host, etc.)
			originalDirector(req)
			// Keep incoming path and raw query, just set scheme/host to backend
			req.URL.Scheme = backend.URL.Scheme
			req.URL.Host = backend.URL.Host
			// Don't blindly change req.Host; preserve original Host header if you want that behavior.
			// But set X-Forwarded-Host so upstream knows original host
			req.Header.Set("X-Forwarded-Host", r.Host)
			req.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
			// Append client IP to X-Forwarded-For
			if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				prior := req.Header.Get("X-Forwarded-For")
				if prior != "" {
					req.Header.Set("X-Forwarded-For", prior+", "+ip)
				} else {
					req.Header.Set("X-Forwarded-For", ip)
				}
			}
			// Optional: set Host header to backend host OR keep original Host.
			// req.Host = backend.URL.Host
		}

		// ModifyResponse allows changing responses before they go to client
		proxy.ModifyResponse = func(resp *http.Response) error {
			// Remove hop-by-hop headers from upstream responses
			removeHopByHopHeaders(resp.Header)

			// Fix Location header if upstream redirects with absolute URL pointing to upstream host.
			if loc := resp.Header.Get("Location"); loc != "" {
				if u, err := url.Parse(loc); err == nil {
					// If redirect points to backend host, rewrite to proxy host
					if u.Host == backend.URL.Host {
						u.Scheme = r.URL.Scheme
						u.Host = r.Host
						resp.Header.Set("Location", u.String())
					}
				}
			}
			return nil
		}

		// Custom error handler for the proxy
		proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
			// Mark backend as dead on errors
			log.Printf("[proxy error] backend %s error: %v", backend.URL, err)
			backend.SetAlive(false)
			http.Error(w, "upstream error", http.StatusBadGateway)
		}

		// Serve using the selected backend's reverse proxy
		proxy.ServeHTTP(w, r)

		dur := time.Since(start)
		log.Printf("[proxy] %s %s -> %s (alive=%v) duration=%s", r.Method, r.URL.Path, backend.URL.String(), backend.IsAlive(), dur)
	})
}

// loggingMiddleware logs basic request info.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Use a ResponseWriter wrapper to capture status if needed (omitted for brevity)
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s in %s", r.RemoteAddr, r.Method, r.URL.String(), time.Since(start))
	})
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		log.Printf("invalid duration for %s: %v, using default %s", key, v, def)
	}
	return def
}

func splitAndTrim(s string, sep string) []string {
	var out []string
	for _, part := range strings.Split(s, sep) {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func Main() {
	// Configuration from env
	listenAddr := getenv("PROXY_LISTEN", ":8080")
	backendCSV := os.Getenv("BACKENDS") // comma-separated list of backend URLs (http://ip:port)
	if backendCSV == "" {
		log.Fatal("BACKENDS environment variable is required (comma-separated urls)")
	}
	backends := splitAndTrim(backendCSV, ",")

	// Timeouts and tuning
	reqTimeout := getenvDuration("PROXY_REQ_TIMEOUT", 60*time.Second) // per request
	healthInterval := getenvDuration("PROXY_HEALTH_INTERVAL", 10*time.Second)
	healthTimeout := getenvDuration("PROXY_HEALTH_TIMEOUT", 2*time.Second)
	healthPath := getenv("PROXY_HEALTH_PATH", "/healthz")

	// Transport tunables
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	pool, err := NewPool(backends, transport)
	if err != nil {
		log.Fatalf("failed to create backend pool: %v", err)
	}

	// Start health checker
	ctx, cancel := context.WithCancel(context.Background())
	go pool.HealthCheck(ctx, healthInterval, healthTimeout, healthPath)

	// Router & handlers
	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// report overall proxy health
		alive := 0
		for _, b := range pool.backends {
			if b.IsAlive() {
				alive++
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "ok\nbackends_alive=%d\nbackends_total=%d\n", alive, len(pool.backends))
	}))
	// simple metrics
	var totalProxied uint64
	mux.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "proxy_total_proxied %d\n", atomic.LoadUint64(&totalProxied))
	}))

	// main proxy
	proxyHandler := ProxyHandler(pool, reqTimeout)
	// Wrap to increment metrics
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&totalProxied, 1)
		proxyHandler.ServeHTTP(w, r)
	}))

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // let proxy transport manage response timeouts
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done
		log.Println("[main] shutting down server")
		// stop health checker
		cancel()
		ctx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[main] graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("[main] starting proxy on %s, backends: %v", listenAddr, backends)
	if strings.HasPrefix(listenAddr, ":") {
		// run plain HTTP
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] ListenAndServe error: %v", err)
		}
	} else {
		// run on explicit addr
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] ListenAndServe error: %v", err)
		}
	}
	log.Println("[main] server stopped")
}
