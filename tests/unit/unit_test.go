package unit

import (
	"testing"
)

// Testa a função de formatação de números de telefone
func TestFormatPhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Valid BR", "5511999999999", "5511999999999"},
		{"With special chars", "(11) 99999-9999", "5511999999999"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPhoneNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FormatPhoneNumber(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Testa validação de API Key
func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		wantValid bool
	}{
		{"Valid Key", "abc123xyz", true},
		{"Empty Key", "", false},
		{"Short Key", "ab", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidateAPIKey(tt.apiKey)
			if valid != tt.wantValid {
				t.Errorf("ValidateAPIKey(%q) = %v; expected %v", tt.apiKey, valid, tt.wantValid)
			}
		})
	}
}

// Funções mock para teste (em produção estariam em outro pacote)
func FormatPhoneNumber(phone string) string {
	if phone == "" {
		return ""
	}
	// Simples remoção de caracteres não numéricos para o teste
	clean := ""
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			clean += string(r)
		}
	}
	if len(clean) == 11 && clean[0] != '5' {
		return "55" + clean
	}
	return clean
}

func ValidateAPIKey(key string) bool {
	return len(key) > 5
}
