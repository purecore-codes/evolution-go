package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Testa a integração completa: HTTP -> Handler -> SQLite -> Response
func TestContactFlowIntegration(t *testing.T) {
	// Setup: Criar DB temporário
	tmpFile, err := os.CreateTemp("", "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Criar tabela
	_, err = db.Exec(`CREATE TABLE contacts (
		id TEXT PRIMARY KEY,
		name TEXT,
		phone TEXT,
		last_seen TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Mock Server simulando API externa
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/contacts" {
			contacts := []map[string]interface{}{
				{"id": "5511999999999", "name": "João Silva", "phone": "5511999999999"},
				{"id": "5511888888888", "name": "Maria Santos", "phone": "5511888888888"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(contacts)
		}
	}))
	defer mockServer.Close()

	// Simular handler que busca contatos e salva no SQLite
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(mockServer.URL + "/contacts")
	if err != nil {
		t.Fatalf("Failed to fetch contacts: %v", err)
	}
	defer resp.Body.Close()

	var contacts []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Salvar no SQLite
	for _, c := range contacts {
		_, err = db.Exec(
			"INSERT INTO contacts (id, name, phone, last_seen) VALUES (?, ?, ?, ?)",
			c["id"], c["name"], c["phone"], time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to insert contact: %v", err)
		}
	}

	// Verificar se os dados foram persistidos
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count contacts: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 contacts, got %d", count)
	}

	// Testar busca
	var name string
	err = db.QueryRow("SELECT name FROM contacts WHERE id = ?", "5511999999999").Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query contact: %v", err)
	}

	if name != "João Silva" {
		t.Errorf("Expected 'João Silva', got %q", name)
	}
}

// Testa integração com contexto de cancelamento
func TestDatabaseContextIntegration(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "test-ctx-*.db")
	defer os.Remove(tmpFile.Name())

	db, err := sql.Open("sqlite3", tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE messages (id INTEGER PRIMARY KEY, content TEXT)")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Inserir dados com contexto
	done := make(chan error)
	go func() {
		_, err := db.ExecContext(ctx, "INSERT INTO messages (content) VALUES (?)", "test message")
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Insert failed: %v", err)
		}
	case <-ctx.Done():
		t.Error("Context timed out unexpectedly")
	}
}
