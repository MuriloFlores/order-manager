# HTTP API Bootstrap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unblock the currently-broken build (missing `entity.Product`, broken `main.go` import), then replace `main.go`'s hardcoded demo scenario with a real Gin HTTP server exposing the existing `OrderController` routes under `/api/v1`.

**Architecture:** Introduce `ports.ProductData` as the DTO Sales uses instead of the never-defined `entity.Product` (fixes `CatalogGateway` and `OrderUseCase.AddOrderItems`, including a live map-indexing regression). Extract HTTP bootstrap into a new `internal/shared/server` package built around a small `RouteRegistrar` interface so future modules can register routes without touching `main.go`. Rewrite `main.go` as a pure composition root: wire dependencies, start the server, read the port from `SERVER_PORT`.

**Tech Stack:** Go 1.26.4, Gin v1.12.0, GORM v1.31.2, `google/uuid` v1.6.0, standard library `testing` (no new dependencies).

**Design doc:** `docs/superpowers/specs/2026-07-04-http-api-bootstrap-design.md`

## Global Constraints

- Go version: 1.26.4 (per `go.mod`) — do not bump.
- No new third-party dependencies. Use the standard `testing` package for all new tests (no testify, no mocking libraries) — hand-written fakes only.
- Code identifiers (variables, structs, functions) in English; comments, if any, may be in Portuguese — matches existing repo convention (`CONTEXT.md` section 9).
- Do not touch `internal/modules/catalog`, `internal/modules/stock`, `internal/modules/delivery`, auth/multi-tenancy, or controller response DTOs — explicitly out of scope per the design doc.
- Follow existing package/file layout conventions (`core/ports`, `use_cases`, `infrastructure/adapters`, `internal/shared/*`) — no unrelated restructuring.

---

### Task 1: `ports.ProductData` DTO + `CatalogGateway` fix

**Files:**
- Modify: `internal/modules/sales/core/ports/product_repository.go`
- Modify: `internal/modules/sales/infrastructure/adapters/catalog_gateway.go`
- Test: `internal/modules/sales/infrastructure/adapters/catalog_gateway_test.go` (new)

**Interfaces:**
- Produces: `ports.ProductData{ ID uuid.UUID; Name string; UnitPrice float64; UnitOfType value_objects.UnitOfType }` (comparable struct — every field is comparable).
- Produces: `ports.ProductRepository.FindAllByIDs(ctx context.Context, productIDs []uuid.UUID) ([]ports.ProductData, error)` — replaces the old `[]entity.Product` return type.
- Produces: `(*adapters.CatalogGateway).FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]ports.ProductData, error)` — same signature, now satisfies `ports.ProductRepository`.
- Consumes: `catalog.CatalogService.GetProductsData(ids []uuid.UUID) ([]catalog.CatalogDTO, error)` (unchanged, already exists in `internal/modules/catalog/public_api.go`).

- [ ] **Step 1: Replace `entity.Product` with `ports.ProductData` in the ports package**

Rewrite `internal/modules/sales/core/ports/product_repository.go` in full:

```go
package ports

import (
	"context"
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

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

This removes the `entity` import from `ports` entirely — `ports` no longer depends on a type that never existed.

- [ ] **Step 2: Write the failing test for `CatalogGateway`**

Create `internal/modules/sales/infrastructure/adapters/catalog_gateway_test.go`:

```go
package adapters

import (
	"context"
	"errors"
	"testing"

	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/core/ports"
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

type stubCatalogService struct {
	data []catalog.CatalogDTO
	err  error
}

func (s *stubCatalogService) GetProductsData(ids []uuid.UUID) ([]catalog.CatalogDTO, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func TestCatalogGateway_FindAllByIDs_ReturnsProductData(t *testing.T) {
	productID := uuid.New()
	stub := &stubCatalogService{
		data: []catalog.CatalogDTO{
			{ID: productID, Name: "Coffee", Price: 12.5},
		},
	}
	gateway := NewCatalogGateway(stub)

	got, err := gateway.FindAllByIDs(context.Background(), []uuid.UUID{productID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 product, got %d", len(got))
	}

	want := ports.ProductData{
		ID:         productID,
		Name:       "Coffee",
		UnitPrice:  12.5,
		UnitOfType: value_objects.Unit,
	}
	if got[0] != want {
		t.Fatalf("got %+v, want %+v", got[0], want)
	}
}

func TestCatalogGateway_FindAllByIDs_PropagatesError(t *testing.T) {
	stub := &stubCatalogService{err: errors.New("catalog unavailable")}
	gateway := NewCatalogGateway(stub)

	_, err := gateway.FindAllByIDs(context.Background(), []uuid.UUID{uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `go test ./internal/modules/sales/infrastructure/adapters/... -run TestCatalogGateway -v`

Expected: FAIL to compile — `catalog_gateway.go` still declares `FindAllByIDs` returning `[]entity.Product`, and `entity.Product` does not exist, so the package fails to build (e.g. `undefined: entity.Product`).

- [ ] **Step 4: Fix `CatalogGateway`**

Rewrite `internal/modules/sales/infrastructure/adapters/catalog_gateway.go` in full:

```go
package adapters

import (
	"context"
	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/core/ports"
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

type CatalogGateway struct {
	catalogAPI catalog.CatalogService
}

func NewCatalogGateway(api catalog.CatalogService) *CatalogGateway {
	return &CatalogGateway{catalogAPI: api}
}

func (g *CatalogGateway) FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]ports.ProductData, error) {
	catalogData, err := g.catalogAPI.GetProductsData(ids)
	if err != nil {
		return nil, err
	}

	var products []ports.ProductData
	for _, item := range catalogData {
		products = append(products, ports.ProductData{
			ID:         item.ID,
			Name:       item.Name,
			UnitPrice:  item.Price,
			UnitOfType: value_objects.Unit,
		})
	}

	return products, nil
}
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/modules/sales/infrastructure/adapters/... -run TestCatalogGateway -v`

Expected: PASS (both `TestCatalogGateway_FindAllByIDs_ReturnsProductData` and `TestCatalogGateway_FindAllByIDs_PropagatesError`).

- [ ] **Step 6: Commit**

```bash
git add internal/modules/sales/core/ports/product_repository.go \
        internal/modules/sales/infrastructure/adapters/catalog_gateway.go \
        internal/modules/sales/infrastructure/adapters/catalog_gateway_test.go
git commit -m "fix: replace entity.Product with ports.ProductData in CatalogGateway"
```

---

### Task 2: Fix `OrderUseCase.AddOrderItems` map-indexing regression

**Files:**
- Modify: `internal/modules/sales/use_cases/order.go`
- Test: `internal/modules/sales/use_cases/order_test.go` (new)

**Interfaces:**
- Consumes: `ports.ProductData` from Task 1 (fields: `ID`, `Name`, `UnitPrice`, `UnitOfType`).
- Consumes: `ports.OrderRepository` (`Save`, `FindByID`, `FindByStatus` — unchanged, from `internal/modules/sales/core/ports/order_repository.go`).
- Consumes: `utils.EventBusInterface` (`Subscribe`, `Publish`, `Wait` — unchanged, from `internal/shared/utils/event_bus.go`).
- Produces: no new exported symbols — this task only fixes the body of the existing `(*OrderUseCase).AddOrderItems(ctx context.Context, orderID uuid.UUID, items map[uuid.UUID]float64) error`.

- [ ] **Step 1: Write the failing regression test**

Create `internal/modules/sales/use_cases/order_test.go`:

```go
package use_cases

import (
	"context"
	"errors"
	"testing"

	"order-manager/internal/modules/sales/core/entity"
	"order-manager/internal/modules/sales/core/ports"
	"order-manager/internal/modules/sales/core/value_objects"
	"order-manager/internal/shared/utils"

	"github.com/google/uuid"
)

type fakeOrderRepo struct {
	order *entity.Order
	saved *entity.Order
}

func (f *fakeOrderRepo) Save(ctx context.Context, order *entity.Order) error {
	f.saved = order
	return nil
}

func (f *fakeOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	if f.order == nil {
		return nil, errors.New("order not found")
	}
	return f.order, nil
}

func (f *fakeOrderRepo) FindByStatus(ctx context.Context, status value_objects.OrderStatus, pagination utils.Pagination) (*utils.PaginatedResult[entity.Order], error) {
	return nil, nil
}

type fakeProductRepo struct {
	products []ports.ProductData
}

func (f *fakeProductRepo) FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]ports.ProductData, error) {
	return f.products, nil
}

type fakeEventBus struct{}

func (f *fakeEventBus) Subscribe(topic string, handler utils.EventHandler) {}
func (f *fakeEventBus) Publish(ctx context.Context, event utils.Event)     {}
func (f *fakeEventBus) Wait()                                             {}

func TestAddOrderItems_MapsQuantityByProductID(t *testing.T) {
	productA := uuid.New()
	productB := uuid.New()

	order := entity.NewOrder("Ana")
	orderRepo := &fakeOrderRepo{order: order}
	productRepo := &fakeProductRepo{
		products: []ports.ProductData{
			{ID: productA, Name: "Coffee", UnitPrice: 10, UnitOfType: value_objects.Unit},
			{ID: productB, Name: "Tea", UnitPrice: 5, UnitOfType: value_objects.Unit},
		},
	}

	uc := NewOrderUseCase(orderRepo, productRepo, &fakeEventBus{})

	items := map[uuid.UUID]float64{
		productA: 3,
		productB: 7,
	}

	if err := uc.AddOrderItems(context.Background(), order.ID, items); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	quantities := make(map[uuid.UUID]float64)
	for _, item := range orderRepo.saved.Items() {
		quantities[item.ProductID] = item.Quantity
	}

	if quantities[productA] != 3 {
		t.Errorf("expected productA quantity 3, got %v", quantities[productA])
	}
	if quantities[productB] != 7 {
		t.Errorf("expected productB quantity 7, got %v", quantities[productB])
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/modules/sales/use_cases/... -run TestAddOrderItems -v`

Expected: FAIL to compile. After Task 1, `products` in `AddOrderItems` is `[]ports.ProductData`, but the current line `qty := items[product]` indexes a `map[uuid.UUID]float64` with a `ports.ProductData` value — a type error (e.g. `cannot use product (variable of struct type ports.ProductData) as uuid.UUID value in map index`).

- [ ] **Step 3: Fix the map index**

In `internal/modules/sales/use_cases/order.go`, inside `AddOrderItems`, change:

```go
	for _, product := range products {
		qty := items[product]
```

to:

```go
	for _, product := range products {
		qty := items[product.ID]
```

No other lines in this function change — `product.ID`, `product.Name`, `product.UnitPrice`, `product.UnitOfType` already match `ports.ProductData`'s field names.

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/modules/sales/use_cases/... -run TestAddOrderItems -v`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/modules/sales/use_cases/order.go internal/modules/sales/use_cases/order_test.go
git commit -m "fix: index items map by product.ID in AddOrderItems"
```

---

### Task 3: `internal/shared/server` package (Gin bootstrap)

**Files:**
- Create: `internal/shared/server/server.go`
- Test: `internal/shared/server/server_test.go` (new)

**Interfaces:**
- Produces: `server.RouteRegistrar` interface with method `RegisterRoutes(rg *gin.RouterGroup)`.
- Produces: `server.New(registrars ...server.RouteRegistrar) *gin.Engine` — builds a `gin.Default()` engine, groups everything under `/api/v1`, and calls `RegisterRoutes` on each registrar with that group.
- Consumes: nothing from this repo besides `github.com/gin-gonic/gin` (already a dependency).

- [ ] **Step 1: Write the failing test**

Create `internal/shared/server/server_test.go`:

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubRegistrar struct{}

func (stubRegistrar) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
}

func TestNew_RegistersRoutesUnderAPIV1(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(stubRegistrar{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "pong" {
		t.Fatalf("expected body %q, got %q", "pong", rec.Body.String())
	}
}

func TestNew_RouteNotAvailableOutsideV1Prefix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := New(stubRegistrar{})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/shared/server/... -v`

Expected: FAIL to compile — `package server` and `New` don't exist yet (`internal/shared/server/server.go` hasn't been created).

- [ ] **Step 3: Implement the server package**

Create `internal/shared/server/server.go`:

```go
package server

import "github.com/gin-gonic/gin"

type RouteRegistrar interface {
	RegisterRoutes(rg *gin.RouterGroup)
}

func New(registrars ...RouteRegistrar) *gin.Engine {
	engine := gin.Default()

	v1 := engine.Group("/api/v1")
	for _, registrar := range registrars {
		registrar.RegisterRoutes(v1)
	}

	return engine
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `go test ./internal/shared/server/... -v`

Expected: PASS (both `TestNew_RegistersRoutesUnderAPIV1` and `TestNew_RouteNotAvailableOutsideV1Prefix`).

- [ ] **Step 5: Commit**

```bash
git add internal/shared/server/server.go internal/shared/server/server_test.go
git commit -m "feat: add internal/shared/server Gin bootstrap package"
```

---

### Task 4: Rewrite `main.go` as composition root + `SERVER_PORT` config

**Files:**
- Modify: `cmd/order-manager/main.go`
- Modify: `.env.example`

**Interfaces:**
- Consumes: `repository.NewPostgresOrderRepo(db *gorm.DB) ports.OrderRepository` (unchanged, from `internal/modules/sales/infrastructure/database/repository/order.go`).
- Consumes: `adapters.NewCatalogGateway(api catalog.CatalogService) *adapters.CatalogGateway` (from Task 1; satisfies `ports.ProductRepository` since `FindAllByIDs` now returns `[]ports.ProductData`).
- Consumes: `use_cases.NewOrderUseCase(orderRepo ports.OrderRepository, productRepo ports.ProductRepository, eventBus utils.EventBusInterface) *use_cases.OrderUseCase` (unchanged).
- Consumes: `controllers.NewOrderController(uc *use_cases.OrderUseCase) *controllers.OrderController`, which already implements `server.RouteRegistrar` via its existing `RegisterRoutes(router *gin.RouterGroup)` method (from Task 3, no change needed to `controllers.go`).
- Consumes: `server.New(registrars ...server.RouteRegistrar) *gin.Engine` (from Task 3).

- [ ] **Step 1: Rewrite `main.go`**

Replace the entire contents of `cmd/order-manager/main.go`:

```go
package main

import (
	"context"
	"log"
	"os"

	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/infrastructure/adapters"
	"order-manager/internal/modules/sales/infrastructure/database/models"
	"order-manager/internal/modules/sales/infrastructure/database/repository"
	"order-manager/internal/modules/sales/infrastructure/http/controllers"
	"order-manager/internal/modules/sales/use_cases"
	"order-manager/internal/shared/database"
	"order-manager/internal/shared/events"
	"order-manager/internal/shared/server"
	"order-manager/internal/shared/utils"
)

func main() {
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("critical database error: %v", err)
	}

	if err := db.AutoMigrate(&models.OrderModel{}, &models.OrderItemModel{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	bus := utils.NewEventBus()

	bus.Subscribe("OrderPaid", func(ctx context.Context, event utils.Event) {
		payload, ok := event.Payload.(events.OrderPaidPayload)
		if !ok {
			log.Println("[Kitchen Module] invalid payload")
			return
		}
		log.Printf("[Kitchen Module] preparing order #%s", payload.OrderID)
		for _, item := range payload.Items {
			log.Printf(" -> %.0fx %s", item.Quantity, item.Name)
		}
	})

	bus.Subscribe("OrderPaid", func(ctx context.Context, event utils.Event) {
		payload, ok := event.Payload.(events.OrderPaidPayload)
		if !ok {
			return
		}
		log.Printf("[Stock Module] deducting ingredients for order #%s", payload.OrderID)
	})

	fakeCatalogModule := catalog.NewFakeCatalogService()
	productGateway := adapters.NewCatalogGateway(fakeCatalogModule)
	orderRepo := repository.NewPostgresOrderRepo(db)
	orderUseCase := use_cases.NewOrderUseCase(orderRepo, productGateway, bus)
	orderController := controllers.NewOrderController(orderUseCase)

	engine := server.New(orderController)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Order Manager listening on :%s", port)
	if err := engine.Run(":" + port); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
```

- [ ] **Step 2: Add `SERVER_PORT` to `.env.example`**

Current `.env.example`:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=order_manager
```

Add a line so it reads:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=order_manager
SERVER_PORT=8080
```

- [ ] **Step 3: Verify the whole module builds**

Run: `go build ./...`

Expected: exits 0, no output (no more `is not in std` import error, no more `undefined: entity.Product`).

- [ ] **Step 4: Run the full test suite**

Run: `go test ./...`

Expected: PASS for `internal/modules/sales/infrastructure/adapters`, `internal/modules/sales/use_cases`, `internal/shared/server`; `ok` (no test files) for every other package.

- [ ] **Step 5: Commit**

```bash
git add cmd/order-manager/main.go .env.example
git commit -m "feat: wire main.go as a composition root running the real Gin server"
```

---

### Task 5: Manual smoke test (end-to-end verification)

**Files:** none (verification only — no commit).

This exercises the full stack for real: Postgres, migrations, the Gin server, and the `OrderPaid` event subscribers, none of which are covered by the unit tests in Tasks 1–3.

- [ ] **Step 1: Start Postgres**

Run: `docker compose up -d db`

Expected: container starts; `docker compose ps` shows the `db` service healthy/running.

- [ ] **Step 2: Start the app**

Run: `go run cmd/order-manager/main.go` (with `.env` present and pointing at the local `db`, per `.env.example`/`docker-compose.yaml`)

Expected: log line `Order Manager listening on :8080`, no fatal errors, process keeps running (foreground).

- [ ] **Step 3: Create an order**

In another terminal:

```bash
curl -s -X POST http://localhost:8080/api/v1/orders/ \
  -H "Content-Type: application/json" \
  -d '{"customer_name":"Ana"}'
```

Expected: HTTP 201, JSON body with `"message":"Order created successfully"` and an `"order_id"` (a UUID). Save it as `ORDER_ID`.

- [ ] **Step 4: Add items (via the fake Catalog)**

```bash
curl -s -X POST http://localhost:8080/api/v1/orders/$ORDER_ID/items/add \
  -H "Content-Type: application/json" \
  -d '{"items_id": {"11111111-1111-1111-1111-111111111111": 2}}'
```

Expected: HTTP 200, `"message":"Order items added successfully"`. (The fake Catalog service returns the same mock product for any UUID.)

- [ ] **Step 5: Pay the order and observe event subscribers**

```bash
curl -s -X POST http://localhost:8080/api/v1/orders/$ORDER_ID/pay
```

Expected: HTTP 200, `"message":"Order paid successfully"`; in the `go run` terminal, log lines from both `[Kitchen Module]` and `[Stock Module]` appear referencing `$ORDER_ID`.

- [ ] **Step 6: Shut down**

Stop the app with Ctrl+C in its terminal, then run: `docker compose down`

Expected: app exits cleanly (no panic); Postgres container stops.
