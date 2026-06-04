# Test Suite em Go - WhatsApp Manager

Este diretório contém uma suíte completa de testes para o sistema de gerenciamento de instâncias WhatsApp.

## 📁 Estrutura de Testes

| Arquivo | Tipo | Descrição |
|---------|------|-----------|
| `unit_test.go` | Unitários | Testa funções isoladas (formatação, validação) |
| `integration_test.go` | Integração | Testa fluxo completo HTTP → SQLite → Response |
| `e2e_test.go` | E2E | Testa o sistema completo em ambiente real |
| `load_test.go` | Load | Simula múltiplos usuários simultâneos |
| `stress_test.go` | Stress | Empurra o sistema além dos limites normais |
| `chaos_test.go` | Chaos | Simula falhas aleatórias e condições adversas |
| `security_test.go` | Security | Testa vulnerabilidades (SQLi, XSS, Auth) |
| `benchmark_test.go` | Benchmark | Mede performance e throughput |

## 🚀 Como Executar

### Pré-requisitos
```bash
cd /workspace/tests
go mod tidy
```

### Variáveis de Ambiente
```bash
export API_BASE_URL="http://localhost:8080"
export API_KEY="sua_api_key_aqui"
export INSTANCE_ID="id_da_instancia"
```

### Executar Todos os Testes
```bash
# Testes unitários e de integração
go test -v ./...

# Com cobertura de código
go test -v -cover ./...

# Apenas testes rápidos (exclui E2E)
go test -v -short ./...
```

### Executar por Categoria

#### Unit Tests
```bash
go test -v -run TestUnit
```

#### Integration Tests
```bash
go test -v -run TestIntegration
```

#### E2E Tests (requer servidor rodando)
```bash
RUN_E2E=1 go test -v -run TestE2E
```

#### Load Tests
```bash
go test -v -run TestLoad -timeout 5m
```

#### Stress Tests
```bash
go test -v -run TestStress -timeout 10m
```

#### Chaos Tests
```bash
go test -v -run TestChaos -timeout 5m
```

#### Security Tests
```bash
go test -v -run TestSecurity
```

#### Benchmarks
```bash
# Executar todos os benchmarks
go test -bench=. -benchmem

# Executar benchmark específico
go test -bench=BenchmarkLoadContacts -benchmem

# Gerar perfil de CPU
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof
```

### Executar com Race Detector
```bash
go test -race -v ./...
```

### Gerar Relatório de Cobertura
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📊 Saída Esperada

### Load Test
```
🚀 Starting Load Test: 50 users, 10 requests each

📊 Load Test Results:
   Total Requests: 500
   Successful: 498
   Failed: 2
   Duration: 15.234s
   Requests/sec: 32.82
```

### Stress Test
```
🔥 Starting Stress Test for 30s with 100 max concurrency

📊 Stress Test Results:
   Duration: 30.012s
   Total Requests: 2847
   Successful: 2801 (98.38%)
   Failed: 46 (1.62%)
   Timeouts: 12
   Avg RPS: 94.86
```

### Security Test
```
🔒 Testing SQL Injection vulnerabilities...
   ✅ Payload blocked/sanitized: ' OR '1'='1
   ✅ Payload blocked/sanitized: '; DROP TABLE contacts; --
✅ No SQL Injection vulnerabilities detected
```

### Benchmark
```
📊 Running Benchmark Suite...
================================

LoadContacts             :      1000 iters |    1.234 ms/op |      12.34 MB/s
SearchContacts           :       856 iters |    1.456 ms/op |      10.21 MB/s
SendMessage              :       654 iters |    2.123 ms/op |       8.45 MB/s

================================
✅ Benchmark suite completed!
```

## ⚠️ Avisos Importantes

1. **E2E Tests**: Requerem o servidor rodando em `API_BASE_URL`
2. **Load/Stress Tests**: Podem sobrecarregar o servidor - use em ambiente de teste
3. **Chaos Tests**: Simulam falhas reais - não execute em produção
4. **Security Tests**: Testam vulnerabilidades - use apenas em ambientes controlados

## 🔧 Troubleshooting

### Erro: "no tests to run"
Verifique se as variáveis de ambiente estão configuradas corretamente.

### Erro: "connection refused"
Certifique-se de que o servidor está rodando no endereço especificado em `API_BASE_URL`.

### Benchmarks lentos
Aumente o timeout: `go test -bench=. -timeout 30m`

### Race conditions detectadas
Execute com race detector: `go test -race ./...`

## 📝 Adicionando Novos Testes

1. Crie um novo arquivo `_test.go` na pasta `tests/`
2. Use o pacote apropriado (`unit`, `integration`, `e2e`, etc.)
3. Siga o padrão `TestXxx(t *testing.T)` para testes
4. Siga o padrão `BenchmarkXxx(b *testing.B)` para benchmarks

## 📈 Métricas de Qualidade

- **Cobertura mínima**: 80%
- **Tempo máximo de teste**: 10 minutos
- **Falhas aceitáveis em chaos/stress**: < 10%
