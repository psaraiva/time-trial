package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:7777/sabotage", "target URL")
	concurrency := flag.Int("c", 10, "concurrent workers")
	total := flag.Int("n", 1000, "total requests")
	flag.Parse()

	var (
		success atomic.Int64
		failure atomic.Int64
		totalNs atomic.Int64
	)

	jobs := make(chan struct{}, *total)
	for i := 0; i < *total; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	client := &http.Client{Timeout: 10 * time.Second}

	start := time.Now()

	var wg sync.WaitGroup
	for range *concurrency {
		wg.Go(func() {
			for range jobs {
				t := time.Now()
				resp, err := client.Get(*url)
				elapsed := time.Since(t).Nanoseconds()
				totalNs.Add(elapsed)

				if err != nil || resp.StatusCode >= 500 {
					failure.Add(1)
				} else {
					success.Add(1)
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		})
	}

	wg.Wait()
	elapsed := time.Since(start)

	total64 := success.Load() + failure.Load()
	avgMs := float64(totalNs.Load()) / float64(total64) / 1e6

	fmt.Printf("\n--- results ---\n")
	fmt.Printf("total:       %d requests in %s\n", total64, elapsed.Round(time.Millisecond))
	fmt.Printf("success:     %d\n", success.Load())
	fmt.Printf("failure:     %d\n", failure.Load())
	fmt.Printf("req/s:       %.2f\n", float64(total64)/elapsed.Seconds())
	fmt.Printf("avg latency: %.2fms\n", avgMs)
}
