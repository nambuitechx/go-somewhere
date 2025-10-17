package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// var allowedCIDRs = []string {
// 	"127.0.0.1", // local
// 	"::1",       // ipv6 local
// 	"0.0.0.0/32",
// }

// func isAllowed(remoteAddr string) bool {
// 	host, _, err := net.SplitHostPort(remoteAddr)
// 	if err != nil {
// 		return false
// 	}

// 	for _, allowed := range allowedCIDRs {
// 		if host == allowed {
// 			return true
// 		}
// 	}

// 	return false
// }

// handleHTTP handles plain HTTP requests
func handleHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	req.RequestURI = "" // Required for client requests
	req.Header.Del("Proxy-Connection")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Connection")

	if !strings.HasPrefix(req.URL.Scheme, "http") {
		http.Error(w, "Unsupported protocol", http.StatusBadRequest)
		return
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		log.Printf("[HTTP] %s %s -> ERROR: %v (%s)", req.Method, req.URL, err, time.Since(start))
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

	log.Printf("[HTTP] %s %s -> %d (%s)", req.Method, req.URL, resp.StatusCode, time.Since(start))
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// handleTunneling handles HTTPS requests using CONNECT
func handleTunneling(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Start bidirectional copy
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)

	log.Printf("[CONNECT] %s OK (%s)", r.Host, time.Since(start))
}

func transfer(dst io.WriteCloser, src io.ReadCloser) {
	defer dst.Close()
	defer src.Close()
	io.Copy(dst, src)
}

func main() {
	server := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// // Optional IP restriction
			// if !isAllowed(r.RemoteAddr) {
			// 	http.Error(w, "Forbidden", http.StatusForbidden)
			// 	log.Printf("[DENY] %s tried to connect", r.RemoteAddr)
			// 	return
			// }

			log.Printf("[Addr] Remote connection address: %s", r.RemoteAddr)
			
			if r.Method == http.MethodConnect {
				handleTunneling(w, r)
			} else {
				handleHTTP(w, r)
			}
		}),
	}

	log.Println("Starting forward proxy on :8888")
	log.Fatal(server.ListenAndServe())
}