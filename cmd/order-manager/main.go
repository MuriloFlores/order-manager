package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/core/entity"
	"order-manager/internal/modules/sales/core/value_objects"
	"order-manager/internal/modules/sales/infrastructure/adapters"
	"order-manager/internal/modules/sales/infrastructure/database"
	"order-manager/internal/modules/sales/infrastructure/database/models"
	"order-manager/internal/modules/sales/use_cases"
	"order-manager/internal/shared/database"
	"order-manager/internal/shared/events"
	"order-manager/internal/shared/utils"
)

func main() {
	fmt.Println("=== Iniciando Order Manager (Real DB + Gateway) ===")
	ctx := context.Background()

	// 1. Inicializar Banco de Dados
	db, err := database.NewPostgresConnection()
	if err != nil {
		fmt.Printf("Erro crítico no banco de dados: %v\n", err)
		return
	}

	// Executa a migração automática das tabelas de infraestrutura
	_ = db.AutoMigrate(&models.OrderModel{}, &models.OrderItemModel{})

	// 2. Inicializar Infraestrutura e EventBus
	bus := utils.NewEventBus()
	
	// Subscribers de Cozinha e Estoque
	bus.Subscribe("OrderPaid", func(ctx context.Context, event utils.Event) {
		payload, ok := event.Payload.(events.OrderPaidPayload)
		if !ok {
			fmt.Printf("[Kitchen Module] ERRO FATAL: Payload inválido\n")
			return
		}
		fmt.Printf("\n--- [KITCHEN MODULE] ---\n")
		fmt.Printf("Recebi o aviso! Preparando pedido #%s\n", payload.OrderID)
		for _, item := range payload.Items {
			fmt.Printf(" -> %.0fx %s\n", item.Quantity, item.Name)
		}
		fmt.Printf("------------------------\n\n")
	})

	bus.Subscribe("OrderPaid", func(ctx context.Context, event utils.Event) {
		payload, ok := event.Payload.(events.OrderPaidPayload)
		if !ok {
			return
		}
		fmt.Printf("[Stock Module] Baixando ingredientes do pedido #%s\n", payload.OrderID)
	})

	// 3. Inicializar Repositórios e Gateways
	
	// O Catálogo Falso que criamos (o outro lado da cerca)
	fakeCatalogModule := catalog.NewFakeCatalogService()
	
	// O Gateway na infraestrutura de Vendas (Nossa ponte para o Catálogo)
	productGateway := adapters.NewCatalogGateway(fakeCatalogModule)
	
	// O Repositório real do Postgres
	postgresOrderRepo := database.NewPostgresOrderRepo(db)

	// 4. Instanciar o Caso de Uso
	orderUseCase := use_cases.NewOrderUseCase(postgresOrderRepo, productGateway, bus)

	// 5. Preparar um cenário de teste real
	fmt.Println("[Main] Testando a Criação de um Pedido Real no Banco...")
	
	// Cria o pedido vazio
	orderID, err := orderUseCase.CreateOrder(ctx, "Cliente VIP")
	if err != nil {
		fmt.Printf("Erro ao criar pedido: %v\n", err)
		return
	}
	fmt.Printf("[Main] Pedido %s criado com sucesso.\n", orderID)

	// Adiciona itens via Gateway
	// Vamos inventar dois UUIDs que o Catálogo Falso vai interceptar
	fakeProductID1 := uuid.New()
	fakeProductID2 := uuid.New()
	
	itemsToAdd := map[uuid.UUID]float64{
		fakeProductID1: 2, // 2x Produto Dinâmico 1
		fakeProductID2: 1, // 1x Produto Dinâmico 2
	}
	
	err = orderUseCase.AddOrderItems(ctx, orderID, itemsToAdd)
	if err != nil {
		fmt.Printf("Erro ao adicionar itens: %v\n", err)
	} else {
		fmt.Println("[Main] Itens adicionados com sucesso (Buscando dados do Catálogo via Gateway).")
	}

	// 6. Pagamento (Dispara Eventos)
	time.Sleep(1 * time.Second) // Apenas para dar tempo de ler os logs
	fmt.Println("[Main] Finalizando a cobrança...")
	
	err = orderUseCase.PayOrder(ctx, orderID)
	if err != nil {
		fmt.Printf("Erro ao pagar pedido: %v\n", err)
	} else {
		fmt.Println("[Main] Pagamento aprovado! Banco atualizado e Eventos publicados.")
	}

	bus.Wait()
	fmt.Println("=== Encerrando ===")
}
