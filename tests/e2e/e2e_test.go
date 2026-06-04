package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Configuração do ambiente de E2E
var baseURL string

func init() {
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
}

// Testa o fluxo completo de ponta a ponta: Instância -> Contatos -> Chat -> Mensagem
func TestFullChatFlowE2E(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		t.Skip("Skipping E2E test. Set RUN_E2E=1 to run.")
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Passo 1: Verificar se há instâncias conectadas
	resp, err := client.Get(baseURL + "/user/instances")
	if err != nil {
		t.Fatalf("Failed to get instances: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var instances []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&instances); err != nil {
		t.Fatalf("Failed to decode instances: %v", err)
	}

	if len(instances) == 0 {
		t.Skip("No instances available for E2E testing")
	}

	instanceID := instances[0]["id"].(string)
	apiKey := instances[0]["api_key"].(string)

	// Passo 2: Carregar contatos da instância
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, instanceID), nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to load contacts: %v", err)
	}
	defer resp.Body.Close()

	var contacts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		t.Fatalf("Failed to decode contacts: %v", err)
	}

	if len(contacts) == 0 {
		t.Skip("No contacts available for E2E testing")
	}

	contactID := contacts[0]["id"].(string)

	// Passo 3: Obter histórico de chat
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/user/chat?instance=%s&contact=%s", baseURL, instanceID, contactID), nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to get chat history: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for chat history, got %d", resp.StatusCode)
	}

	// Passo 4: Enviar mensagem
	msgPayload := map[string]string{
		"to":      contactID,
		"message": "Teste E2E automatizado",
	}
	payload, _ := json.Marshal(msgPayload)

	req, _ = http.NewRequest("POST", fmt.Sprintf("%s/user/send?instance=%s", baseURL, instanceID), bytes.NewBuffer(payload))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Failed to send message: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var sendResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sendResp)

	if sendResp["success"] != true {
		t.Error("Message send reported failure")
	}
}

// Testa recuperação de erro e resiliência
func TestErrorRecoveryE2E(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		t.Skip("Skipping E2E test. Set RUN_E2E=1 to run.")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Tentar acessar com API Key inválida
	req, _ := http.NewRequest("GET", baseURL+"/user/contacts?instance=invalid", nil)
	req.Header.Set("Authorization", "Bearer invalid_key")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Deve retornar 401 ou 403
	if resp.StatusCode != http.StatusUnauthorized && resp.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 401 or 403, got %d", resp.StatusCode)
	}
}
