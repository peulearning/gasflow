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
