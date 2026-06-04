package stress

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Teste de Stress: Empurra o sistema além dos limites normais
func TestStressContactsEndpoint(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "test-instance-001"
	}

	duration := 30 * time.Second
	maxConcurrency := 100

	var wg sync.WaitGroup
	var totalRequests int64
	var successRequests int64
	var failedRequests int64
	var timeoutRequests int64

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	startTime := time.Now()
	fmt.Printf("🔥 Starting Stress Test for %v with %d max concurrency\n", duration, maxConcurrency)

	semaphore := make(chan struct{}, maxConcurrency)

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		semaphore <- struct{}{}
		wg.Add(1)

		go func(reqNum int) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			atomic.AddInt64(&totalRequests, 1)

			client := &http.Client{Timeout: 5 * time.Second}
			req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err := client.Do(req)

			if err != nil {
				if os.IsTimeout(err) || (resp != nil && resp.StatusCode == http.StatusGatewayTimeout) {
					atomic.AddInt64(&timeoutRequests, 1)
				} else {
					atomic.AddInt64(&failedRequests, 1)
				}
			} else if resp.StatusCode == http.StatusOK {
				atomic.AddInt64(&successRequests, 1)
			} else {
				atomic.AddInt64(&failedRequests, 1)
			}

			if resp != nil {
				resp.Body.Close()
			}

			if reqNum%100 == 0 {
				fmt.Printf("   [%d] Success: %d, Failed: %d, Timeouts: %d\n",
					reqNum, successRequests, failedRequests, timeoutRequests)
			}
		}(int(totalRequests))
	}

done:
	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Printf("\n📊 Stress Test Results:\n")
	fmt.Printf("   Duration: %v\n", elapsed)
	fmt.Printf("   Total Requests: %d\n", totalRequests)
	fmt.Printf("   Successful: %d (%.2f%%)\n", successRequests, float64(successRequests)/float64(totalRequests)*100)
	fmt.Printf("   Failed: %d (%.2f%%)\n", failedRequests, float64(failedRequests)/float64(totalRequests)*100)
	fmt.Printf("   Timeouts: %d\n", timeoutRequests)
	fmt.Printf("   Avg RPS: %.2f\n", float64(totalRequests)/elapsed.Seconds())

	if float64(failedRequests)/float64(totalRequests) > 0.1 {
		t.Errorf("High failure rate: %.2f%%", float64(failedRequests)/float64(totalRequests)*100)
	}
}

// Teste de Stress com pico súbito (Spike Test)
func TestStressSpikeLoad(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "test-instance-001"
	}

	spikeSize := 200
	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	fmt.Printf("⚡ Starting Spike Test with %d concurrent requests\n", spikeSize)

	startTime := time.Now()

	for i := 0; i < spikeSize; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}

			req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err := client.Do(req)
			mu.Lock()
			if err != nil || resp.StatusCode != http.StatusOK {
				failCount++
			} else {
				successCount++
			}
			mu.Unlock()

			if resp != nil {
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Printf("\n📊 Spike Test Results:\n")
	fmt.Printf("   Concurrent Requests: %d\n", spikeSize)
	fmt.Printf("   Duration: %v\n", elapsed)
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", failCount)

	if failCount > spikeSize/10 {
		t.Errorf("Spike test had high failure rate: %d/%d", failCount, spikeSize)
	}
}
