package benchmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	baseURL    string
	apiKey     string
	instanceID string
	httpClient *http.Client
)

func init() {
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey = os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	instanceID = os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "test-instance-001"
	}

	httpClient = &http.Client{Timeout: 30 * time.Second}
}

// Benchmark para carregamento de contatos
func BenchmarkLoadContacts(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// Benchmark para busca de contatos com filtro
func BenchmarkSearchContacts(b *testing.B) {
	searchTerm := "João"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s&search=%s", baseURL, instanceID, searchTerm), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// Benchmark para envio de mensagem
func BenchmarkSendMessage(b *testing.B) {
	msgPayload := map[string]string{
		"to":      "5511999999999",
		"message": "Benchmark test message",
	}
	payload, _ := json.Marshal(msgPayload)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/user/send?instance=%s", baseURL, instanceID), bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			// Não falhar o benchmark por erros de rede/API externa
			continue
		}
		resp.Body.Close()
	}
}

// Benchmark para obtenção de histórico de chat
func BenchmarkGetChatHistory(b *testing.B) {
	contactID := "5511999999999"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/chat?instance=%s&contact=%s", baseURL, instanceID, contactID), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
	}
}

// Benchmark para listagem de instâncias
func BenchmarkListInstances(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", baseURL+"/user/instances", nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := httpClient.Do(req)
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}
}

// Benchmark completo de fluxo (carregar contatos + buscar chat)
func BenchmarkFullChatFlow(b *testing.B) {
	contactID := "5511999999999"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Passo 1: Carregar contatos
		req1, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
		req1.Header.Set("Authorization", "Bearer "+apiKey)
		resp1, err := httpClient.Do(req1)
		if err != nil {
			continue
		}
		resp1.Body.Close()

		// Passo 2: Buscar chat
		req2, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/chat?instance=%s&contact=%s", baseURL, instanceID, contactID), nil)
		req2.Header.Set("Authorization", "Bearer "+apiKey)
		resp2, err := httpClient.Do(req2)
		if err != nil {
			continue
		}
		resp2.Body.Close()
	}
}

// Benchmark concorrente para carregamento de contatos
func BenchmarkLoadContactsConcurrent(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
			req.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err := httpClient.Do(req)
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			resp.Body.Close()
		}
	})
}

// Benchmark com relatório detalhado
func BenchmarkReport(b *testing.B) {
	fmt.Println("\n📊 Running Benchmark Suite...")
	fmt.Println("================================")

	benchmarks := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"LoadContacts", BenchmarkLoadContacts},
		{"SearchContacts", BenchmarkSearchContacts},
		{"SendMessage", BenchmarkSendMessage},
		{"GetChatHistory", BenchmarkGetChatHistory},
		{"ListInstances", BenchmarkListInstances},
		{"FullChatFlow", BenchmarkFullChatFlow},
		{"LoadContactsConcurrent", BenchmarkLoadContactsConcurrent},
	}

	for _, bm := range benchmarks {
		result := testing.Benchmark(bm.fn)
		fmt.Printf("\n%-25s: %10d iters | %12v/op\n",
			bm.name,
			result.N,
			result.NsPerOp(),
		)
	}

	fmt.Println("\n================================")
	fmt.Println("✅ Benchmark suite completed!")
}
