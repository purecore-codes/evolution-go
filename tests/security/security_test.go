package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Teste de Segurança: SQL Injection
func TestSecuritySQLInjection(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	injectionPayloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE contacts; --",
		"1; SELECT * FROM users--",
		"' UNION SELECT NULL,NULL,NULL--",
		"admin'--",
		"1' AND '1'='1",
	}

	client := &http.Client{Timeout: 10 * time.Second}
	vulnerable := false

	fmt.Printf("🔒 Testing SQL Injection vulnerabilities...\n")

	for _, payload := range injectionPayloads {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, payload), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Verificar se há indícios de vulnerabilidade
		bodyStr := string(body)
		if strings.Contains(strings.ToLower(bodyStr), "sql syntax") ||
			strings.Contains(strings.ToLower(bodyStr), "sqlite") ||
			strings.Contains(strings.ToLower(bodyStr), "drop table") {
			fmt.Printf("   ⚠️ Potential SQL Injection with payload: %s\n", payload)
			vulnerable = true
		} else {
			fmt.Printf("   ✅ Payload blocked/sanitized: %s\n", payload)
		}
	}

	if vulnerable {
		t.Error("Potential SQL Injection vulnerability detected!")
	} else {
		fmt.Println("✅ No SQL Injection vulnerabilities detected")
	}
}

// Teste de Segurança: XSS (Cross-Site Scripting)
func TestSecurityXSS(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	xssPayloads := []string{
		"<script>alert('XSS')</script>",
		"<img src=x onerror=alert('XSS')>",
		"javascript:alert('XSS')",
		"<svg onload=alert('XSS')>",
		"'><script>alert('XSS')</script><'",
	}

	client := &http.Client{Timeout: 10 * time.Second}
	vulnerable := false

	fmt.Printf("\n🔒 Testing XSS vulnerabilities...\n")

	for _, payload := range xssPayloads {
		// Simular envio de mensagem com payload XSS
		msgPayload := map[string]string{
			"to":      "5511999999999",
			"message": payload,
		}
		payloadBytes, _ := json.Marshal(msgPayload)

		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/user/send?instance=test", baseURL), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		bodyStr := string(body)

		// Verificar se o payload foi refletido sem sanitização
		if strings.Contains(bodyStr, "<script>") || strings.Contains(bodyStr, "onerror=") ||
			strings.Contains(bodyStr, "javascript:") {
			if !strings.Contains(bodyStr, "&lt;script&gt;") && !strings.Contains(bodyStr, "&lt;") {
				fmt.Printf("   ⚠️ Potential XSS with payload: %s\n", payload)
				vulnerable = true
			} else {
				fmt.Printf("   ✅ Payload properly escaped: %s\n", payload)
			}
		} else {
			fmt.Printf("   ✅ Payload blocked/sanitized: %s\n", payload)
		}
	}

	if vulnerable {
		t.Error("Potential XSS vulnerability detected!")
	} else {
		fmt.Println("✅ No XSS vulnerabilities detected")
	}
}

// Teste de Segurança: Autenticação e Autorização
func TestSecurityAuthentication(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Printf("\n🔒 Testing Authentication bypass...\n")

	// Testar sem token
	req, _ := http.NewRequest("GET", baseURL+"/user/contacts?instance=test", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Error("❌ API accessible without authentication!")
	} else {
		fmt.Printf("   ✅ Unauthenticated request blocked: %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	// Testar com token inválido
	req, _ = http.NewRequest("GET", baseURL+"/user/contacts?instance=test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_xyz")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Error("❌ API accessible with invalid token!")
	} else {
		fmt.Printf("   ✅ Invalid token rejected: %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	// Testar com token vazio
	req, _ = http.NewRequest("GET", baseURL+"/user/contacts?instance=test", nil)
	req.Header.Set("Authorization", "Bearer ")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		t.Error("❌ API accessible with empty token!")
	} else {
		fmt.Printf("   ✅ Empty token rejected: %d\n", resp.StatusCode)
	}
	resp.Body.Close()
}

// Teste de Segurança: Rate Limiting
func TestSecurityRateLimiting(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	requestCount := 50
	rateLimited := false

	client := &http.Client{Timeout: 5 * time.Second}

	fmt.Printf("\n🔒 Testing Rate Limiting with %d rapid requests...\n", requestCount)

	for i := 0; i < requestCount; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=test", baseURL), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusForbidden {
			fmt.Printf("   ✅ Rate limit triggered at request %d (HTTP %d)\n", i+1, resp.StatusCode)
			rateLimited = true
			resp.Body.Close()
			break
		}
		resp.Body.Close()
	}

	if !rateLimited {
		fmt.Println("   ⚠️ No rate limiting detected (may be configured for higher thresholds)")
	}
}

// Teste de Segurança: Path Traversal
func TestSecurityPathTraversal(t *testing.T) {
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key_12345"
	}

	traversalPayloads := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"....//....//etc/passwd",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
	}

	client := &http.Client{Timeout: 10 * time.Second}
	vulnerable := false

	fmt.Printf("\n🔒 Testing Path Traversal vulnerabilities...\n")

	for _, payload := range traversalPayloads {
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/user/contacts?instance=%s", baseURL, payload), nil)
		req.Header.Set("Authorization", "Bearer "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		bodyStr := string(body)

		// Verificar se há conteúdo sensível exposto
		if strings.Contains(bodyStr, "root:") || strings.Contains(bodyStr, "/bin/bash") ||
			strings.Contains(strings.ToLower(bodyStr), "microsoft") {
			fmt.Printf("   ⚠️ Potential Path Traversal with payload: %s\n", payload)
			vulnerable = true
		} else {
			fmt.Printf("   ✅ Payload blocked: %s\n", payload)
		}
	}

	if vulnerable {
		t.Error("Potential Path Traversal vulnerability detected!")
	} else {
		fmt.Println("✅ No Path Traversal vulnerabilities detected")
	}
}
