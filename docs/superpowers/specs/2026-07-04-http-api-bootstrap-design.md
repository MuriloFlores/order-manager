# Design: Subir a API HTTP real (Gin) e resolver dívida bloqueadora

**Data:** 2026-07-04
**Módulo:** Sales (+ shared)
**Status:** Aprovado, aguardando plano de implementação

## 1. Motivação

O `main.go` atual roda um cenário de teste hardcoded (cria pedido, adiciona itens, paga) em vez de subir um servidor HTTP real, mesmo o `OrderController` (com todas as rotas REST) já existindo em `infrastructure/http/controllers/controllers.go`.

Além disso, o projeto **não compila hoje**, por dois motivos que bloqueiam diretamente esse trabalho:

1. `main.go` importa `order-manager/internal/modules/sales/infrastructure/database`, pacote que não existe (a implementação real do repositório está em `.../infrastructure/database/repository`, pacote `repository`). Resquício de uma reestruturação anterior.
2. `ports.ProductRepository` e `adapters.CatalogGateway` referenciam `entity.Product`, um tipo que nunca foi definido no código. Esta é a "Pendência Crítica" já documentada no `CONTEXT.md` (seção 6), com solução acordada mas não implementada: um DTO `ports.ProductData`.

Há também uma regressão não commitada em `use_cases/order.go` (`AddOrderItems`): `qty := items[product.ID]` foi alterado para `qty := items[product]`, incompatível com o tipo do map (`map[uuid.UUID]float64`). Corrigida como efeito colateral do item 2, ao reescrever esse trecho para usar o DTO.

## 2. Escopo

**Dentro do escopo:**
- Corrigir o import quebrado do `main.go`.
- Implementar `ports.ProductData` e remover toda referência a `entity.Product` do módulo Sales.
- Extrair o bootstrap HTTP para um pacote dedicado `internal/shared/server`.
- Reescrever `main.go` como composition root puro: monta dependências e sobe o servidor. Remove o cenário de teste hardcoded.
- Registrar as rotas do `OrderController` sob o prefixo `/api/v1`.
- Configurar a porta do servidor via variável de ambiente `SERVER_PORT` (default `8080`), adicionada ao `.env.example`.

**Fora do escopo (dívida conhecida, não deste design):**
- Graceful shutdown do servidor.
- DTOs de resposta dedicados nos controllers (continuam usando `gin.H{}`).
- Autenticação / multi-tenancy.
- Módulos de Catalog real, Stock, Delivery.

## 3. Correção da dívida bloqueadora

### 3.1 `ports.ProductData` (substitui `entity.Product`)

`ports/product_repository.go` passa a definir um DTO puro (sem regra de negócio, pertence ao vocabulário de Sales — não ao de Catalog):

```go
type ProductData struct {
    ID         uuid.UUID
    Name       string
    UnitPrice  float64
    UnitOfType value_objects.UnitOfType
}

type ProductRepository interface {
    FindAllByIDs(ctx context.Context, productIDs []uuid.UUID) ([]ProductData, error)
}
```

Isso remove a dependência de `ports` em `entity`.

`adapters/catalog_gateway.go`: `FindAllByIDs` passa a montar e devolver `[]ports.ProductData` em vez de `[]entity.Product`.

`use_cases/order.go`: `AddOrderItems` passa a iterar `[]ports.ProductData`:

```go
for _, product := range products {
    qty := items[product.ID]
    item, err := entity.NewOrderItem(product.ID, product.Name, product.UnitPrice, qty, product.UnitOfType)
    ...
}
```

Isso corrige de graça a regressão do map (`items[product]` → `items[product.ID]`).

### 3.2 Import quebrado do `main.go`

Troca o import inválido `order-manager/internal/modules/sales/infrastructure/database` pelo pacote correto `order-manager/internal/modules/sales/infrastructure/database/repository`, e usa `repository.NewPostgresOrderRepo(db)`.

## 4. Pacote de servidor HTTP

Novo arquivo `internal/shared/server/server.go`, `package server`:

```go
type RouteRegistrar interface {
    RegisterRoutes(rg *gin.RouterGroup)
}

func New(registrars ...RouteRegistrar) *gin.Engine {
    engine := gin.Default()
    v1 := engine.Group("/api/v1")
    for _, r := range registrars {
        r.RegisterRoutes(v1)
    }
    return engine
}
```

`*controllers.OrderController` já satisfaz `RegisterRoutes(rg *gin.RouterGroup)` sem alterações — a interface só formaliza um contrato que já existe. Isso permite que futuros controllers (Stock, Delivery) sejam adicionados à lista de `registrars` sem alterar o pacote `server`.

## 5. `main.go` — novo fluxo (composition root)

1. Conectar no Postgres (`database.NewPostgresConnection`). Em caso de erro: `log.Fatalf` (em vez do atual `fmt.Printf` + `return` silencioso), garantindo que o processo saia com código de erro — relevante porque o `docker-compose.yaml` usa `restart: on-failure`.
2. `db.AutoMigrate(&models.OrderModel{}, &models.OrderItemModel{})`.
3. Montar: `EventBus` — **mantendo os dois `bus.Subscribe("OrderPaid", ...)` de demonstração (Kitchen/Stock)**, que continuam válidos: são handlers de evento, não parte do cenário de teste, e agora reagem a pagamentos reais feitos via HTTP —, `catalog.FakeCatalogService`, `adapters.CatalogGateway`, `repository.NewPostgresOrderRepo`, `use_cases.NewOrderUseCase`, `controllers.NewOrderController`.
4. `engine := server.New(orderController)`.
5. Ler `SERVER_PORT` do ambiente (`os.Getenv`, default `"8080"` se vazio) e chamar `engine.Run(":" + port)`; erro de `Run` também vira `log.Fatalf`.

O que é **removido** é só a parte "cenário de teste": as chamadas diretas a `CreateOrder`/`AddOrderItems`/`PayOrder` no `main`, e o `time.Sleep` + `bus.Wait()` usados para esperar os logs do cenário hardcoded. Os endpoints passam a ser testados via HTTP real (curl/Postman) a partir de agora, e os handlers de Kitchen/Stock continuam imprimindo no console quando um pedido é pago pela API.

## 6. Configuração

`.env.example` ganha a linha `SERVER_PORT=8080`, consistente com o mapeamento `"8080:8080"` já existente em `docker-compose.yaml`.

## 7. Testes / Verificação

- `go build ./...` deve compilar sem erros (verifica a correção dos dois bloqueadores).
- Smoke test manual via HTTP: subir o servidor (`go run cmd/order-manager/main.go` com Postgres rodando via `docker compose up -d db`) e exercitar o fluxo completo via curl: `POST /api/v1/orders/` → `POST /api/v1/orders/:id/items/add` → `POST /api/v1/orders/:id/pay`, conferindo que o EventBus ainda dispara os subscribers de Kitchen/Stock (que continuam registrados em `main.go`, só que agora ouvindo eventos reais do servidor, não do cenário de teste).

## 8. Arquivos afetados

- `internal/modules/sales/core/ports/product_repository.go`
- `internal/modules/sales/infrastructure/adapters/catalog_gateway.go`
- `internal/modules/sales/use_cases/order.go`
- `cmd/order-manager/main.go`
- `internal/shared/server/server.go` (novo)
- `.env.example`
