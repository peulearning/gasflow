# 🚀 GasFlow

**GasFlow** é um sistema completo de gestão, vendas e distribuição. Construído sob uma arquitetura de **Monólito Modular**, o sistema utiliza mensageria assíncrona para garantir alta performance e resiliência, separando responsabilidades de forma clara entre domínios.

## 🛠️ Stack Tecnológico

* **Backend:** Go (Chi Router)
* **Frontend:** Next.js (React)
* **Banco de Dados:** MySQL 8
* **Mensageria:** RabbitMQ
* **Infraestrutura:** Docker & Docker Compose

---

## 🏗️ Arquitetura

O backend em Go segue os princípios de **Domain-Driven Design (DDD)** e **Clean Architecture** dentro de um modelo de Monólito Modular.

* `internal/domain`: Regras de negócio puras, entidades, Value Objects e Máquina de Estados.
* `internal/modules`: Casos de uso específicos por contexto (`clients`, `orders`, `inventory`, `billing`, `analytics`).
* `internal/infra`: Implementações técnicas (MySQL, RabbitMQ, JWT).

### Máquina de Estados (Pedidos)
O ciclo de vida do pedido é estritamente controlado por uma FSM (Finite State Machine) validada no domínio:

```text
[received] ➔ [approved] ➔ [separated] ➔ [in_route] ➔ [delivered]
      ↳ [cancelled]             ↳ [cancelled]       ↳ [rescheduled] ➔ [in_route]

```

---

## 🚀 Como Executar LocalmenteO projeto já está configurado com Docker para facilitar o setup do ambiente de desenvolvimento.Clone o repositório:

Bash
git clone [https://github.com/seu-usuario/gasflow.git](https://github.com/seu-usuario/gasflow.git)

cd gasflow

Configure as variáveis de ambiente baseadas no .env.example (crie um .env na raiz).Suba os containers (API, Frontend, MySQL e RabbitMQ):

Bash
docker-compose up -d

Acesse as aplicações:
API: http://localhost:8080
Frontend: http://localhost:3000
RabbitMQ Management: http://localhost:15672 (Credenciais no .env)

📡 Eventos e Mensageria (RabbitMQ)O sistema utiliza arquitetura orientada a eventos para desacoplar módulos. Cada fila possui uma DLQ (Dead-Letter Queue) associada para retentativas exponenciais em caso de falha.


Exchange | Routing Key | Produtor | Consumidores | Ação
orders | order.created,Orders| "Inventory | Billing" | Notifica novo pedido
orders | order.status_changed | Orders |"Analytics , Billing" | Atualiza read-models e faturamento
orders | order.delivered | Orders | "Inv, Bill , Analytics" | Confirma baixa de estoque e gera cobrança
inventory | stock.reserved | Inventory | Orders | Confirma reserva física
inventory | stock.released | Inventory | Orders | Libera reserva (cancelamento)
billing | charge.generated | Billing | Analytics |Atualiza receita no dashboard


## 🛣️ Rotas Principais da API

A API é construída de forma RESTful, com autenticação via JWT.

Autenticação: POST /api/auth/login | POST /api/auth/refresh

Clientes: GET /api/clients | POST /api/clients | GET /api/clients/:id/orders

Pedidos: POST /api/orders | PATCH /api/orders/:id/status (Transição FSM)

Estoque: POST /api/inventory/deposits/:id/receive | GET /api/inventory/low-stock

Faturamento: GET /api/charges/overdue

Dashboard: GET /api/dashboard/kpis (Polling a cada 30s pelo frontend)


##🧪 Estratégia de Testes
O projeto utiliza um pipeline rigoroso de testes, categorizado pelo Makefile:

Bash
# Testes unitários (rápidos, regras de negócio puras e FSM)
make test-unit

# Testes de integração (usa testcontainers-go para MySQL e RabbitMQ)
make test-integration

# Testes End-to-End (E2E - fluxo completo via chamadas HTTP)
make test-e2e

# Rodar a suíte completa
make test-all


🗺️ Roadmap
Fase 1 — MVP (4~6 semanas) 🚀

[ ] Auth JWT com roles (Admin, Operational, Financial).

[ ] CRUD base (Clientes, Produtos, Motoristas).

[ ] Pedidos com FSM de status e reserva de estoque síncrona.

[ ] Dashboard integrado com 5 KPIs essenciais via polling.

Fase 2 — Distribuição Real (3~4 semanas) 📦

[ ] Implementação total dos eventos RabbitMQ.

[ ] Read-model analytics_daily populado via consumers.

[ ] Implementação de DLQ e retry exponencial.

[ ] Testes de integração robustos com testcontainers-go.

Fase 3 — Qualidade e Produção (3~4 semanas) 🛡️

[ ] Tabela de auditoria (audit_logs) monitorando todas as ações críticas.

[ ] Logs estruturados (Ex: zerolog).

[ ] Exposição de métricas no endpoint /metrics (Prometheus-ready).

[ ] Cobertura de testes ≥ 80% nos módulos core.