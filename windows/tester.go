package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TestEndpoints provides IPv6 test endpoints
var TestEndpoints = []string{
	"https://api64.ipify.org",
	"https://v6.ident.me",
	"https://ipv6.icanhazip.com",
}

// TestResult holds the result of testing a single IPv6
type TestResult struct {
	IP       net.IP
	Success  bool
	PublicIP string
	Latency  time.Duration
	Error    error
}

// TestIPv6Connectivity tests outbound connection for a given local IPv6 address
func TestIPv6Connectivity(localIP net.IP, endpoint string) TestResult {
	start := time.Now()
	res := TestResult{IP: localIP}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := net.Dialer{
				LocalAddr: &net.TCPAddr{IP: localIP},
				Timeout:   5 * time.Second,
			}
			return dialer.DialContext(ctx, network, addr)
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   6 * time.Second,
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		res.Latency = time.Since(start)
		res.Error = err
		return res
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		res.Latency = time.Since(start)
		res.Error = err
		return res
	}

	res.Latency = time.Since(start)
	res.Success = true
	res.PublicIP = strings.TrimSpace(string(body))
	return res
}

// BatchTestIPv6 tests a sample of IPv6 addresses concurrently
func BatchTestIPv6(ips []net.IP, maxSample int, onResult func(TestResult)) []TestResult {
	if len(ips) == 0 {
		return nil
	}

	sample := ips
	if len(sample) > maxSample {
		sample = sample[:maxSample]
	}

	var results []TestResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, 5) // max 5 concurrent tests

	for _, ip := range sample {
		wg.Add(1)
		go func(targetIP net.IP) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Try primary endpoint
			res := TestIPv6Connectivity(targetIP, TestEndpoints[0])
			if !res.Success && len(TestEndpoints) > 1 {
				// Fallback
				res = TestIPv6Connectivity(targetIP, TestEndpoints[1])
			}

			mu.Lock()
			results = append(results, res)
			if onResult != nil {
				onResult(res)
			}
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	return results
}
