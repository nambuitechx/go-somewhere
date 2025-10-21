package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"
)

// List of backend servers
var backends = []*url.URL{}

// Counter for round-robin
var counter int32

func mustParse(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatal(err)
	}

	return u
}

// Round-robin load balancer handler
func loadBalancer(w http.ResponseWriter, r *http.Request) {
	// Pick the next backend atomically
	idx := atomic.AddInt32(&counter, 1)
	target := backends[idx % int32(len(backends))]

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Optionally modify the request before proxying
	r.Host = target.Host

	fmt.Printf("Proxying request to %s\n", target)
	proxy.ServeHTTP(w, r)
}

func main() {
	// Init backends
	numberOfBackendsString := os.Getenv("NUMBER_OF_BACKENDS")
	numberOfBackends, err := strconv.ParseInt(numberOfBackendsString, 10, 64)
	if err != nil {
		log.Fatalln("Failed to parse number of backends")
	}

	for i := 0; i < int(numberOfBackends); i++ {
		backends = append(backends, mustParse(fmt.Sprintf("http://backend%d:800%d", i, i)))
	}

	// LB
	http.HandleFunc("/", loadBalancer)

	fmt.Println("Starting round-robin load balancer on :8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
