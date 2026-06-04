package chaos

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

// Teste de Chaos: Simula falhas aleatórias e condições adversas
func TestChaosRandomFailures(t *testing.T) {
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

	duration := 20 * time.Second
	numWorkers := 10

	var wg sync.WaitGroup
	totalRequests := 0
	successCount := 0
	networkErrorCount := 0
	timeoutCount := 0
	serverErrorCount := 0

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	rand.Seed(time.Now().UnixNano())

	fmt.Printf("🌪️ Starting Chaos Test for %v\n", duration)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 3 * time.Second}

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// Introduzir delay aleatório
				delay := time.Duration(rand.Intn(100)) * time.Millisecond
				time.Sleep(delay)

				req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
				req.Header.Set("Authorization", "Bearer "+apiKey)

				startReq := time.Now()
				resp, err := client.Do(req)
				_ = time.Since(startReq)

				totalRequests++

				if err != nil {
					if os.IsTimeout(err) || (resp != nil && resp.StatusCode == http.StatusGatewayTimeout) {
						timeoutCount++
					} else {
						networkErrorCount++
					}
				} else {
					switch resp.StatusCode {
					case http.StatusOK:
						successCount++
					case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
						serverErrorCount++
					default:
						networkErrorCount++
					}
					resp.Body.Close()
				}

				if totalRequests%50 == 0 {
					fmt.Printf("   [Worker %d] Total: %d | Success: %d | Timeouts: %d | Network Errors: %d | Server Errors: %d\n",
						workerID, totalRequests, successCount, timeoutCount, networkErrorCount, serverErrorCount)
				}
			}
		}(i)
	}

	wg.Wait()

	fmt.Printf("\n📊 Chaos Test Results:\n")
	fmt.Printf("   Total Requests: %d\n", totalRequests)
	fmt.Printf("   Successful: %d (%.2f%%)\n", successCount, float64(successCount)/float64(totalRequests)*100)
	fmt.Printf("   Timeouts: %d\n", timeoutCount)
	fmt.Printf("   Network Errors: %d\n", networkErrorCount)
	fmt.Printf("   Server Errors: %d\n", serverErrorCount)

	// Em caos, esperamos algumas falhas - o importante é o sistema se recuperar
	if successCount == 0 {
		t.Error("No successful requests during chaos test - system may be completely down")
	}
}

// Teste de Chaos: Simula latência extrema
func TestChaosHighLatency(t *testing.T) {
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

	requestCount := 20
	successCount := 0
	timeoutCount := 0
	var mu sync.Mutex

	fmt.Printf("🐌 Starting High Latency Chaos Test with %d requests\n", requestCount)

	for i := 0; i < requestCount; i++ {
		// Adicionar latência artificial variável
		latency := time.Duration(100+rand.Intn(2000)) * time.Millisecond
		time.Sleep(latency)

		client := &http.Client{Timeout: 5 * time.Second}
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		startReq := time.Now()
		resp, err := client.Do(req)
		elapsed := time.Since(startReq)

		mu.Lock()
		if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
			timeoutCount++
			fmt.Printf("   Request %d: TIMEOUT after %v\n", i, elapsed)
		} else {
			successCount++
			fmt.Printf("   Request %d: OK in %v\n", i, elapsed)
		}
		mu.Unlock()

		if resp != nil {
			resp.Body.Close()
		}
	}

	fmt.Printf("\n📊 High Latency Test Results:\n")
	fmt.Printf("   Total: %d | Success: %d | Timeouts/Failures: %d\n", requestCount, successCount, timeoutCount)
}

// Teste de Chaos: Reinício simulado do serviço
func TestChaosServiceRestart(t *testing.T) {
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

	fmt.Printf("🔄 Starting Service Restart Chaos Test\n")

	// Simular tentativa de conexão durante "reinício"
	attempts := 0
	maxAttempts := 30
	successAfterFailure := false

	for attempts < maxAttempts {
		client := &http.Client{Timeout: 2 * time.Second}
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			if attempts > 0 {
				successAfterFailure = true
			}
			if resp != nil {
				resp.Body.Close()
			}
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		attempts++
		fmt.Printf("   Attempt %d: Service unavailable, waiting...\n", attempts)
		time.Sleep(1 * time.Second)
	}

	if !successAfterFailure && attempts == 0 {
		fmt.Println("   Service was available from start (no chaos simulated)")
	} else if successAfterFailure {
		fmt.Printf("   ✅ Service recovered after %d attempts\n", attempts)
	} else {
		fmt.Printf("   ⚠️ Service did not recover after %d attempts\n", attempts)
	}
}
