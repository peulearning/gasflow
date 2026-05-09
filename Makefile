# ─────────────────────────────────────────────────────────────────────────────
help: ## Mostra ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n",$$1,$$2}'

# ── Docker ────────────────────────────────────────────────────────────────────
up: ## Sobe todos os containers (build + start)
	docker compose up -d --build

down: ## Para os containers
	docker compose down

down-v: ## Para containers e apaga volumes (reset total)
	docker compose down -v

logs: ## Acompanha logs da API
	docker compose logs -f api

logs-all: ## Acompanha logs de todos os serviços
	docker compose logs -f

# ── Dev local (infra no Docker, API no host) ──────────────────────────────────
dev: ## Sobe MySQL+RabbitMQ e roda API localmente
	docker compose up -d mysql rabbitmq
	@echo "⏳ Aguardando MySQL e RabbitMQ..."
	@sleep 5
	ENV=development \
	DB_HOST=localhost DB_PORT=3306 DB_USER=gasflow DB_PASSWORD=gasflow DB_NAME=gasflow \
	RABBITMQ_URL=amqp://guest:guest@localhost:5672/ \
	JWT_SECRET=dev-secret-troque-em-prod \
	DEFAULT_DEPOSIT_ID=dep-sp-001 \
	go run ./cmd/api

# ── Build ─────────────────────────────────────────────────────────────────────
build: ## Compila o binário
	@mkdir -p bin
	go build -ldflags="-s -w" -o ./bin/gasflow ./cmd/api
	@echo "✅ Binário: ./bin/gasflow"

# ── Testes ────────────────────────────────────────────────────────────────────
test: ## Testes unitários do domínio (sem I/O)
	go test -race -count=1 ./internal/domain/...

test-v: ## Testes com saída detalhada
	go test -race -count=1 -v ./internal/domain/...

test-cover: ## Relatório de cobertura HTML
	@mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./internal/domain/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "✅ Abra: coverage/coverage.html"

# ── Qualidade ─────────────────────────────────────────────────────────────────
fmt: ## Formata o código
	go fmt ./...

vet: ## Roda go vet
	go vet ./...

# ── Utilitários ───────────────────────────────────────────────────────────────
clean: ## Remove artefatos de build
	rm -rf ./bin ./coverage

deps: ## Baixa dependências
	go mod download && go mod tidy

rabbit-ui: ## Abre RabbitMQ Management UI
	open http://localhost:15672 2>/dev/null || xdg-open http://localhost:15672 2>/dev/null || \
		echo "Acesse: http://localhost:15672 (guest/guest)"

# ── Informação ────────────────────────────────────────────────────────────────
info: ## Mostra URLs e credenciais
	@echo ""
	@echo "  🔥 GasFlow — URLs"
	@echo "  API:       http://localhost:8080"
	@echo "  Health:    http://localhost:8080/health"
	@echo "  RabbitMQ:  http://localhost:15672  (guest/guest)"
	@echo "  MySQL:     localhost:3306  (gasflow/gasflow)"
	@echo ""
	@echo "  👤 Usuários de teste (senha: password)"
	@echo "  admin@gasflow.com     → admin"
	@echo "  operador@gasflow.com  → operational"
	@echo "  financeiro@gasflow.com → financial"
	@echo ""
MAKE