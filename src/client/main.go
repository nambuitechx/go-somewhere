package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

var success, fail uint32

func worker(wg *sync.WaitGroup, client *http.Client, target *url.URL, jobs <-chan int) {
	defer wg.Done()

	for range jobs {
		resp, err := client.Get(target.String())
		if err != nil {
			log.Printf("request failed: %s", err.Error())
			atomic.AddUint32(&fail, 1)
			continue
		}

		_ = resp.Body.Close()
		atomic.AddUint32(&success, 1)

		// if i % 100 == 0 {
		// 	log.Printf("Job %d is done!", i)
		// }
	}
}

func main()  {
	var targetString string
	var concurrency int
	var total int

	flag.StringVar(&targetString, "target", "http://localhost:8888", "Target URL string")
	flag.IntVar(&concurrency, "con", 10, "Number of concurrent workers")
	flag.IntVar(&total, "total", 100, "Total number of requests")
	flag.Parse()

	target, err := url.Parse(targetString)
	if err != nil {
		log.Fatal("Target string is invalid")
	}

	wg := &sync.WaitGroup{}
	jobs := make(chan int)

	client := &http.Client {
		Timeout: time.Second * 20,
		Transport: &http.Transport{
			MaxIdleConns: concurrency,
			MaxIdleConnsPerHost: concurrency,
			IdleConnTimeout: time.Second * 60,
			DisableKeepAlives: false,
		},
	}

	log.Printf("Start doing load test on target %s with %d workers and %d total request!\n", targetString, concurrency, total)
	start := time.Now()

	for i := 1; i <= concurrency; i++ {
		wg.Add(1)
		go worker(wg, client, target, jobs)
	}

	for i := 1; i <= total; i++ {
		jobs <- i
	}

	close((jobs))

	wg.Wait()
	elapsed := time.Since(start)

	log.Printf("Done %d load test request on target %s in %.2f seconds!\n", total, targetString, elapsed.Seconds())
	log.Printf("Success: %d -- Fail: %d\n", success, fail)
}
