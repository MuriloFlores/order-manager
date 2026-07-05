# 🏪 Order Manager

**Sistema de gerenciamento de estoque, vendas e entregas para lojas físicas e digitais.**

Um monolito modular construído em Go, projetado para escalar para uma plataforma SaaS multi-tenant — onde lojistas poderão criar contas, cadastrar múltiplas lojas e filiais, e orquestrar toda a operação a partir de um único painel.

---

## 🎯 A Visão

O Order Manager nasceu como um projeto de estudo aprofundado de **Domain-Driven Design (DDD)** e **Clean Architecture**, mas com uma ambição real: se tornar uma plataforma SaaS completa para o varejo(Um sonho um tanto quanto impossível, mas sonhar é grátis).

### Onde Queremos Chegar

```
┌───────────────────────────────────────────────────────────────┐
│                      PLATAFORMA SAAS                          │
│                                                               │
│  Conta do Lojista                                             │ 
│  ├── 🏬 Loja Centro          ├── 🏬 Loja Shopping            │
│  │   ├── 📦 Estoque          │   ├── 📦 Estoque              │
│  │   ├── 🛒 Vendas           │   ├── 🛒 Vendas               │
│  │   └── 🚚 Entregas         │   └── 🚚 Entregas             │
│  │                           │                                │
│  └── Painel Administrativo (Multi-loja)                       │
└───────────────────────────────────────────────────────────────┘
```

- **Multi-Tenant:** Cada lojista terá sua conta isolada com dados segregados.
- **Multi-Loja:** Dentro de uma conta, o lojista poderá gerenciar diversas lojas e filiais.
- **Módulo de Vendas:** Criação, pagamento e cancelamento de pedidos.
- **Módulo de Estoque:** Controle de produtos, categorias, preços e quantidades.
- **Módulo de Delivery:** Rastreio de entregas físicas e encomendas com motoboys.
- **Módulo de Catálogo:** Fonte de verdade sobre produtos, consumido pelos demais módulos.

### Onde Estamos Hoje

O projeto está na fase de construção do **núcleo de domínio do módulo de Vendas**, com foco em solidificar a arquitetura antes de expandir. Atualmente temos:

- [ ] Módulo de Ventas
    - [x] Entidades de Domínio (`Order`, `OrderItem`) com regras de negócio encapsuladas
    - [x] Value Objects (`OrderStatus`, `UnitOfType`)
    - [x] Portas de saída (Interfaces de Repositório)
    - [x] Caso de Uso (`OrderUseCase`) com fluxo completo de criação, adição de itens, pagamento e cancelamento
    - [x] Repositório PostgreSQL com GORM (mapeamento Entity ↔ Model)
    - [x] EventBus para comunicação assíncrona entre módulos
    - [x] Anti-Corruption Layer (Gateway) para comunicação com o Catálogo
    - [ ] Endpoints HTTP (REST API com Gin)
- [ ] Módulo de Catálogo real (com persistência própria)
- [ ] Módulo de Estoque
- [ ] Módulo de Delivery
- [ ] Autenticação e Multi-Tenancy
- [x] Docker Compose com PostgreSQL


---

## 🏛️ Arquitetura

O sistema segue o padrão de **Monolito Modular** com **Clean Architecture (Onion)** e **DDD (Domain-Driven Design)**. Cada módulo é isolado como se fosse um microsserviço, mas roda no mesmo processo — facilitando o desenvolvimento inicial e permitindo uma futura migração para microsserviços sem reescrever o domínio.

```
internal/
├── modules/
│   ├── sales/                          # Módulo de Vendas
│   │   ├── core/                       # 🟢 Camada de Domínio (Zero dependências externas)
│   │   │   ├── entity/                 #    Entidades e Agregados (Order, OrderItem)
│   │   │   ├── value_objects/          #    Objetos de Valor (OrderStatus, UnitOfType)
│   │   │   └── ports/                  #    Interfaces de saída (Repository contracts)
│   │   ├── use_cases/                  # 🔵 Camada de Aplicação (Orquestração)
│   │   └── infrastructure/             # 🔴 Camada de Infraestrutura
│   │       ├── adapters/               #    Gateways (Anti-Corruption Layer)
│   │       ├── database/
│   │       │   ├── models/             #    Structs GORM (schema do banco)
│   │       │   └── repository/         #    Implementações dos Repositórios
│   │       └── http/
│   │           └── controllers/        #    Handlers HTTP (Gin)
│   ├── catalog/                        # Módulo de Catálogo (Mock)
│   ├── stock/                          # Módulo de Estoque (Futuro)
│   └── delivery/                       # Módulo de Delivery (Futuro)
├── shared/                             # Código compartilhado entre módulos
│   ├── database/                       #    Factory de conexão PostgreSQL
│   ├── events/                         #    Payloads de eventos de domínio
│   └── utils/                          #    EventBus, Paginação
└── cmd/
    └── order-manager/
        └── main.go                     # Entrypoint (Injeção de Dependências)
```

### Princípios Respeitados

| Princípio | Como é aplicado |
|-----------|----------------|
| **Isolamento de Entidades** | Entidades no `core/` nunca possuem tags de banco (`gorm`, `json`). Structs separados em `models/` fazem a tradução. |
| **Anti-Corruption Layer** | Módulos se comunicam via Gateways que traduzem DTOs externos para entidades internas. |
| **Linguagem Ubíqua** | No Catálogo é `Product`. Em Vendas é `OrderItem`. Cada contexto tem seu próprio vocabulário. |
| **Dependency Inversion** | O `core/` define interfaces (`ports/`). A infraestrutura as implementa. O domínio nunca depende de framework. |

---

## 🛠️ Tecnologias

| Tecnologia | Versão | Uso |
|------------|--------|-----|
| **Go** | 1.26 | Linguagem principal |
| **Gin** | 1.12 | Framework HTTP (REST API) |
| **GORM** | 1.31 | ORM para PostgreSQL |
| **PostgreSQL** | 15 (Alpine) | Banco de dados relacional |
| **Docker / Docker Compose** | - | Containerização e orquestração local |
| **UUID** | v1.6 | Identificadores únicos de entidades |

---

## 🚀 Como Rodar

### Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)
- [Docker e Docker Compose](https://docs.docker.com/get-docker/)

### Setup

```bash
# 1. Clone o repositório
git clone https://github.com/MuriloFlores/order-manager.git
cd order-manager

# 2. Configure as variáveis de ambiente
cp .env.example .env
# Edite o .env com suas credenciais se necessário

# 3. Suba o banco de dados
docker compose up -d db

# 4. Instale as dependências e rode
go mod tidy
go run cmd/order-manager/main.go
```

### Com Docker (Aplicação completa)

```bash
docker compose up --build
```

---

## 📄 Licença

Este projeto está sob desenvolvimento ativo e ainda não possui uma licença definida.

---

<p align="center">
  Construído com ☕ e <b>Go</b> — Arquitetura limpa, sem atalhos.
</p>
