package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/infrastructure/adapters"
	"order-manager/internal/modules/sales/infrastructure/database/models"
	"order-manager/internal/modules/sales/infrastructure/database/repository"
	"order-manager/internal/modules/sales/infrastructure/http/controllers"
	"order-manager/internal/modules/sales/use_cases"
	"order-manager/internal/shared/database"
	"order-manager/internal/shared/server"
	"order-manager/internal/shared/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"
)

func main() {
	db, err := database.NewPostgresConnection()
	if err != nil {
		fmt.Printf("Erro crítico no banco de dados: %v\n", err)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	_ = db.AutoMigrate(&models.OrderModel{}, &models.OrderItemModel{})
	bus := utils.NewEventBus()
	fakeCatalogModule := catalog.NewFakeCatalogService()
	productGateway := adapters.NewCatalogGateway(fakeCatalogModule)
	postgresOrderRepo := repository.NewPostgresOrderRepo(db)
	orderUC := use_cases.NewOrderUseCase(postgresOrderRepo, productGateway, bus)
	orderCtrl := controllers.NewOrderController(orderUC)

	router := server.RegisterRouter(orderCtrl)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Criando na porta %v", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	GracefulShutdown(srv, db)
}

func GracefulShutdown(srv *http.Server, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}

	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			log.Println("Closing database connection...")
			_ = sqlDB.Close()
		}
	}

	log.Println("Server gracefully stopped")
}
