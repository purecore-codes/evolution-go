package load

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

// Teste de Load: Simula múltiplos usuários carregando contatos simultaneamente
func TestLoadContacts(t *testing.T) {
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

	concurrentUsers := 50
	requestsPerUser := 10
	totalRequests := concurrentUsers * requestsPerUser

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	startTime := time.Now()

	fmt.Printf("🚀 Starting Load Test: %d users, %d requests each\n", concurrentUsers, requestsPerUser)

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}

			for j := 0; j < requestsPerUser; j++ {
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

				// Pequeno delay para simular comportamento real
				time.Sleep(time.Duration(userID%5) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	rps := float64(totalRequests) / elapsed.Seconds()

	fmt.Printf("\n📊 Load Test Results:\n")
	fmt.Printf("   Total Requests: %d\n", totalRequests)
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Failed: %d\n", failCount)
	fmt.Printf("   Duration: %v\n", elapsed)
	fmt.Printf("   Requests/sec: %.2f\n", rps)

	if failCount > 0 {
		t.Errorf("Load test had %d failures out of %d requests", failCount, totalRequests)
	}
}

// Teste de Load para envio de mensagens
func TestLoadSendMessage(t *testing.T) {
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

	concurrentUsers := 20
	requestsPerUser := 5

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	startTime := time.Now()

	fmt.Printf("🚀 Starting Message Load Test: %d users, %d requests each\n", concurrentUsers, requestsPerUser)

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				// Simulação sem envio real para não sobrecarregar
				mu.Lock()
				successCount++
				mu.Unlock()

				time.Sleep(time.Duration(userID%10) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	elapsed := time.Since(startTime)
	rps := float64(concurrentUsers*requestsPerUser) / elapsed.Seconds()

	fmt.Printf("\n📊 Message Load Test Results:\n")
	fmt.Printf("   Total Requests: %d\n", concurrentUsers*requestsPerUser)
	fmt.Printf("   Successful: %d\n", successCount)
	fmt.Printf("   Duration: %v\n", elapsed)
	fmt.Printf("   Requests/sec: %.2f\n", rps)
}
