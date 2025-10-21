package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL 	*url.URL
	Proxy	*httputil.ReverseProxy
	Alive	bool
	mu		sync.RWMutex
}

func (b *Backend) IsAlive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.Alive
}

func (b * Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

type LoadBalancer struct {
	Backends	[]*Backend
	Counter		uint32
}

func NewLoadBalancer(targets []*url.URL) *LoadBalancer {
	lb := &LoadBalancer{}

	for _, t := range targets {
		lb.Backends = append(lb.Backends, &Backend{
			URL: t,
			Proxy: httputil.NewSingleHostReverseProxy(t),
			Alive: false,
		})
	}

	return lb
}

func (lb *LoadBalancer) getNextPeer() *Backend {
	total := len(lb.Backends)
	
	for i := 0; i < total; i++ {
		idx := atomic.AddUint32(&lb.Counter, 1) % uint32(total)
		backend := lb.Backends[idx]

		if backend.IsAlive() {
			return backend
		}
	}

	return nil
}

func (lb *LoadBalancer) healthCheck(client *http.Client, interval time.Duration) {
	for {
		allHealthy := true

		for _, backend := range lb.Backends {
			status := isBackendAlive(client, backend.URL)
			backend.SetAlive(status)

			if !status {
				allHealthy = false
			}
		}

		if allHealthy {
			log.Println("All target servers are healthy")
		}

		time.Sleep(interval)
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.getNextPeer()
	if backend == nil {
		http.Error(w, "No healthy backend available", http.StatusServiceUnavailable)
		return
	}

	r.Host = backend.URL.Host
	log.Printf("Proxying request to %s", backend.URL.String())
	backend.Proxy.ServeHTTP(w, r)
}

func isBackendAlive(client *http.Client, target *url.URL) bool {
	resp, err := client.Get(target.String())

	if err != nil {
		log.Printf("Target %s is unhealthy", target.String())
		return false
	}

	_ = resp.Body.Close()
	return true
}

func main()  {
	// Server
	portString := os.Getenv("PORT")
	if portString == "" {
		portString = "8888"
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		log.Fatalln("PORT env is invalid")
	}

	// Init backends
	backendsString := os.Getenv("BACKENDS")
	backends := strings.Split(backendsString, ",")
	targets := []*url.URL{}

	for _, backend := range backends {
		target, err := url.Parse(backend)

		if err != nil {
			log.Fatalln("Failed to load backend targets")
		}

		targets = append(targets, target)
	}

	// Init lb
	lb := NewLoadBalancer(targets)

	go func ()  {
		client := &http.Client{
			Timeout: time.Second * 10,
		}

		lb.healthCheck(client, time.Second * 5)
	}()

	// HTTP Server
	http.Handle("/", lb)

	log.Printf("Proxy server is running on port %s", portString)
	
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalf("Error in starting server: %s", err.Error())
	}
}
