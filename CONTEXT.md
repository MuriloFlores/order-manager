# CONTEXT.md — Guia Completo para IAs e Colaboradores

> Este arquivo é a **fonte de verdade** sobre o projeto Order Manager.
> Se você é uma IA (ChatGPT, Claude, Gemini, Copilot, Antigravity, etc.) ou um desenvolvedor humano
> lendo este projeto pela primeira vez, **leia este documento inteiro antes de sugerir qualquer código.**

---

## 1. Sobre o Projeto

**Order Manager** é um sistema de gerenciamento de **estoque, vendas e entregas** para lojas físicas e digitais.
Atualmente é um **Monolito Modular** escrito em **Go**, mas a visão de longo prazo é se tornar uma
**plataforma SaaS multi-tenant**, onde lojistas criam contas, cadastram múltiplas lojas/filiais e
orquestram toda a operação.

### Módulos Planejados

| Módulo       | Bounded Context       | Status               |
|--------------|-----------------------|----------------------|
| **Sales**    | Vendas e Pedidos      | 🟡 Em desenvolvimento |
| **Catalog**  | Produtos e Categorias | 🔴 Mock (FakeCatalogService) |
| **Stock**    | Estoque e Inventário  | 🔴 Não iniciado      |
| **Delivery** | Entregas e Motoboys   | 🔴 Não iniciado      |
| **Auth**     | Autenticação / Multi-Tenant | 🔴 Não iniciado |

---

## 2. Dinâmica de Trabalho (Ensino Socrático)

**REGRA FUNDAMENTAL:** O dono deste projeto é um desenvolvedor em processo de estudo de DDD e Clean Architecture.
A IA atua exclusivamente como **Mentor Técnico**, seguindo o método socrático:

- ✅ **PODE:** Propor planos, revisar código, guiar a arquitetura, mostrar exemplos em blocos Markdown, fazer perguntas provocativas.
- ❌ **NÃO PODE:** Implementar ou escrever código diretamente nos arquivos do projeto (a menos que o usuário peça explicitamente naquela interação, por exemplo, tarefas de infraestrutura como `.env`, `docker-compose`, `README`, etc.).
- ✅ **PODE:** Implementar arquivos que o usuário considere "fora do escopo de estudo" (configuração, CI/CD, Docker, README, `.gitignore`, etc.), desde que autorizado.

**Formato de resposta esperado para tarefas de código:**
1. **O Diagnóstico:** Explique o que está certo e errado no código do usuário.
2. **O Desafio:** Liste os pontos de correção numerados.
3. **Material de Apoio:** Mostre snippets de código em blocos Markdown como referência, nunca escrevendo direto no arquivo.

---

## 3. Filosofia Arquitetural

### 3.1 Monolito Modular

O sistema é um **Monolito Modular** — cada módulo é isolado como se fosse um microsserviço, mas roda no mesmo
processo Go. Os módulos compartilham o mesmo binário e banco de dados físico, mas cada um possui suas próprias
tabelas e **nunca** acessa diretamente os dados de outro módulo.

**Por que não microsserviços?** Porque estamos em fase de estudo e desenvolvimento inicial. O monolito modular
permite iterar rápido sem a complexidade de rede, service discovery e deploy distribuído. Quando o sistema
amadurecer, a migração para microsserviços será cirúrgica (basta trocar a chamada in-process por HTTP/gRPC).

### 3.2 Clean Architecture (Onion Architecture)

Cada módulo segue estritamente a arquitetura em camadas concêntricas:

```
┌─────────────────────────────────────────────┐
│           infrastructure/                    │  Camada Suja (Frameworks, DB, HTTP)
│  ┌───────────────────────────────────────┐  │
│  │          use_cases/                    │  │  Camada de Aplicação (Orquestração)
│  │  ┌─────────────────────────────────┐  │  │
│  │  │           core/                  │  │  │  Camada de Domínio (Regras Puras)
│  │  │  ┌───────────────────────────┐  │  │  │
│  │  │  │   entity/ + value_objects/ │  │  │  │  O Coração: ZERO dependências externas
│  │  │  └───────────────────────────┘  │  │  │
│  │  │           ports/                 │  │  │  Contratos (Interfaces de Saída)
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**Regra de Dependência:** As setas de importação apontam SEMPRE para dentro. A infraestrutura depende do core.
O core NUNCA importa a infraestrutura.

### 3.3 Domain-Driven Design (DDD)

- **Aggregate Root:** `Order` é a raiz de agregação. Todo acesso a `OrderItem` passa pelo `Order`.
- **Value Objects:** `OrderStatus` e `UnitOfType` são imutáveis e auto-validantes.
- **Factory Methods:** `NewOrder()` para criação com regras de negócio. `RestoreOrder()` para reidratação do banco.
- **Ubiquitous Language:** No Catálogo é `Product`. Em Vendas é `OrderItem`. Cada contexto tem seu vocabulário.

---

## 4. Regras Invioláveis de Código

### 4.1 Isolamento de Entidades (Impedance Mismatch)

**PROIBIDO:** Tags de banco de dados (`gorm`, `json`, `bson`) em qualquer struct dentro de `core/entity/`.

As entidades são puras. A tradução entre o mundo OO (entidades) e o mundo relacional (tabelas) é feita
por structs separados em `infrastructure/database/models/`, que possuem funções de mapeamento:
- `OrderToModel(order *entity.Order) OrderModel` → Converte entidade para model GORM.
- `OrderItemToModel(item entity.OrderItem) OrderItemModel` → Converte item (sem OrderID, que é injetado pelo pai).
- `RestoreOrder(...)` / `RestoreOrderItem(...)` → Factory methods para reconstruir entidades a partir do banco.

**Detalhe crítico:** O campo `OrderID` (chave estrangeira) existe APENAS no `OrderItemModel` (GORM).
A entidade `OrderItem` não sabe que bancos relacionais existem. Quem injeta o `OrderID` no model é a função
`OrderToModel`, que tem acesso ao ID do pedido pai durante o mapeamento.

### 4.2 Comunicação Inter-Módulos (Anti-Corruption Layer)

**PROIBIDO:** Um módulo acessar diretamente o banco de dados ou as entidades internas de outro módulo.

A comunicação segue o padrão **Gateway / Anti-Corruption Layer (ACL)**:

```
[Sales UseCase] → [ports.ProductRepository] → [CatalogGateway] → [catalog.CatalogService] → [Catalog Module]
                    (Interface no core)         (Adapter na infra)    (Interface pública)        (Implementação)
```

1. O módulo **provedor** (Catalog) expõe uma **Interface Pública** (`CatalogService`) que retorna DTOs burros (`CatalogDTO`).
2. O módulo **consumidor** (Sales) define uma **Porta de Saída** (`ports.ProductRepository`) no seu core.
3. O **Gateway** (`CatalogGateway`) vive na infraestrutura do consumidor, implementa a porta, chama a API pública do provedor e **traduz** o DTO externo para os tipos internos de Sales.

Este padrão é idêntico ao usado em microsserviços — a única diferença é que a "rede" é uma chamada de função in-process.

### 4.3 Linguagem Ubíqua

| Conceito         | No Catálogo   | Em Vendas       | No Banco (GORM)     |
|------------------|---------------|-----------------|---------------------|
| Item à venda     | `Product`     | `OrderItem`     | `OrderItemModel`    |
| Pedido           | —             | `Order`         | `OrderModel`        |
| ID do produto    | `Product.ID`  | `OrderItem.ProductID` | `OrderItemModel.ProductID` |
| ID do pedido     | —             | `Order.ID`      | `OrderModel.ID` / `OrderItemModel.OrderID` |

---

## 5. Estrutura de Pastas (Estado Atual)

```
order-manager/
├── cmd/
│   └── order-manager/
│       └── main.go                             # Entrypoint: Injeção de dependências
├── internal/
│   ├── modules/
│   │   ├── catalog/                            # Módulo de Catálogo
│   │   │   └── public_api.go                   #   Interface pública + FakeCatalogService (Mock)
│   │   └── sales/                              # Módulo de Vendas
│   │       ├── core/                           #   🟢 Domínio Puro
│   │       │   ├── entity/
│   │       │   │   ├── order.go                #     Aggregate Root (NewOrder, RestoreOrder, Pay, Cancel, AddItem)
│   │       │   │   └── order_item.go           #     Entidade filha (NewOrderItem, RestoreOrderItem, SubTotal)
│   │       │   ├── value_objects/
│   │       │   │   ├── order_status.go         #     Enum: Pending, Paid, Cancelled
│   │       │   │   └── unit_of_type.go         #     Enum: Unit, Kg, L
│   │       │   └── ports/
│   │       │       ├── order_repository.go     #     Interface: Save, FindByID, FindByStatus
│   │       │       └── product_repository.go   #     Interface: FindAllByIDs (⚠️ pendente: retorna entity.Product)
│   │       ├── use_cases/
│   │       │   └── order.go                    #   🔵 Casos de uso: Create, AddItems, Pay, Cancel, List
│   │       └── infrastructure/                 #   🔴 Implementações
│   │           ├── adapters/
│   │           │   └── catalog_gateway.go      #     ACL: Traduz CatalogDTO → entity.Product
│   │           ├── database/
│   │           │   ├── models/
│   │           │   │   ├── order.go            #     GORM Model + OrderToModel()
│   │           │   │   └── order_item.go       #     GORM Model + OrderItemToModel()
│   │           │   └── repository/
│   │           │       └── order.go            #     PostgresOrderRepo (Save, FindByID, FindByStatus)
│   │           └── http/
│   │               └── controllers/
│   │                   └── controllers.go      #     Gin handlers (Create, Pay, Cancel, AddItems, RemoveItems, List)
│   └── shared/                                 # Código compartilhado
│       ├── database/
│       │   └── postgres.go                     #   Factory de conexão GORM (lê variáveis de ambiente)
│       ├── events/
│       │   └── order_paid_payload.go           #   Payload do evento OrderPaid
│       └── utils/
│           ├── event_bus.go                    #   EventBus in-memory (Publish/Subscribe)
│           └── paginated.go                    #   Structs de paginação genérica
├── .env.example                                # Template de variáveis de ambiente
├── docker-compose.yaml                         # PostgreSQL 15 + App (lê do .env)
├── Dockerfile                                  # Multi-stage build (Go Alpine)
└── README.md                                   # Documentação pública do projeto
```

---

## 6. Pendências e Débitos Técnicos Conhecidos

### ⚠️ Pendência Crítica: `entity.Product` em Sales

A interface `ports.ProductRepository` retorna `[]entity.Product`, mas esse tipo **não deveria existir** em Sales.
`Product` pertence ao bounded context de **Catalog**, não de Sales.

**Solução acordada (ainda não implementada):**
Criar um struct `ProductData` dentro do pacote `ports` de Sales. É um DTO puro (sem regras de negócio) que
representa os dados que o Gateway precisa devolver. A interface passaria a retornar `[]ports.ProductData`.
Arquivos afetados: `ports/product_repository.go`, `adapters/catalog_gateway.go`, `use_cases/order.go`.

### ⚠️ main.go ainda no modo "script de teste"

O `main.go` atualmente roda um cenário hardcoded de teste (cria pedido, adiciona itens, paga).
Ele precisa ser refatorado para iniciar o servidor HTTP (Gin) e registrar as rotas do `OrderController`.

### ⚠️ Controllers sem DTOs de Response

Os controllers retornam `gin.H{}` diretamente. Idealmente, devem usar DTOs de Response dedicados para
não vazar detalhes internos das entidades.

---

## 7. Decisões Arquiteturais Tomadas (ADRs Informais)

| Decisão | Motivo |
|---------|--------|
| **GORM como ORM** | Facilita o desenvolvimento inicial. O usuário não quer ficar preso no banco. |
| **EventBus in-memory** | Suficiente para o monolito. Pode ser trocado por RabbitMQ/Kafka no futuro. |
| **FakeCatalogService** | O módulo de Catálogo ainda não existe de verdade. O mock permite desenvolver Sales isoladamente. |
| **RestoreOrder / RestoreOrderItem** | Factory methods "backdoor" para reidratar entidades do banco sem disparar regras de negócio. |
| **OrderID só no Model (GORM)** | A entidade `OrderItem` não precisa saber quem é o pai. O mapeador `OrderToModel` injeta a FK. |
| **ProductID na entidade OrderItem** | O `OrderItem` precisa referenciar qual produto do catálogo ele representa, mas apenas pelo UUID. |
| **Campos privados na entidade Order** | `status`, `totalValue`, `createdAt` e `items` são privados. Acesso via getters. Mutação apenas via métodos de negócio. |

---

## 8. Tecnologias

| Tecnologia       | Versão   | Uso                              |
|------------------|----------|----------------------------------|
| Go               | 1.26     | Linguagem principal              |
| Gin              | 1.12     | Framework HTTP (REST API)        |
| GORM             | 1.31     | ORM para PostgreSQL              |
| PostgreSQL       | 15       | Banco de dados relacional        |
| Docker Compose   | 3.8      | Orquestração de containers       |
| UUID (google)    | 1.6      | Identificadores de entidades     |

---

## 9. Como Contribuir (Para IAs e Humanos)

1. **Leia este arquivo inteiro** antes de fazer qualquer sugestão.
2. **Respeite as camadas.** Nunca importe `infrastructure` dentro de `core`.
3. **Respeite os bounded contexts.** `Product` é do Catálogo. `OrderItem` é de Vendas.
4. **Nunca coloque tags de framework em entidades.** Use Models na infraestrutura.
5. **Siga o padrão Gateway/ACL** para qualquer comunicação entre módulos.
6. **Use inglês no código** (nomes de variáveis, structs, funções). Comentários podem ser em português.
7. **Prefira o método socrático.** Guie, não implemente (a menos que explicitamente autorizado).
