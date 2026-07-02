package use_cases

import (
	"context"
	"errors"
	"order-manager/internal/modules/sales/core/entity"
	"order-manager/internal/modules/sales/core/ports"
	"order-manager/internal/modules/sales/core/value_objects"
	"order-manager/internal/shared/events"
	"order-manager/internal/shared/utils"

	"github.com/google/uuid"
)

var (
	ErrEmptyCustomerName = errors.New("empty customer name")
	ErrEmptyItems        = errors.New("empty items")
	ErrNoOrderFound      = errors.New("no order found")
)

type OrderUseCase struct {
	orderRepo   ports.OrderRepository
	productRepo ports.ProductRepository
	eventBus    utils.EventBusInterface
}

func NewOrderUseCase(orderRepo ports.OrderRepository, productRepo ports.ProductRepository, eventBus utils.EventBusInterface) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		eventBus:    eventBus,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, customerName string) (uuid.UUID, error) {
	if customerName == "" {
		return uuid.Nil, ErrEmptyCustomerName
	}

	order := entity.NewOrder(customerName)

	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return uuid.Nil, err
	}

	uc.eventBus.Publish(ctx, utils.Event{
		Name:    "OrderCreated",
		Payload: order.ID.String(),
	})

	return order.ID, nil
}

func (uc *OrderUseCase) AddOrderItems(ctx context.Context, orderID uuid.UUID, items map[uuid.UUID]float64) error {
	if len(items) == 0 {
		return ErrEmptyItems
	}

	var productIDs []uuid.UUID
	for id := range items {
		productIDs = append(productIDs, id)
	}

	products, err := uc.productRepo.FindAllByIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	for _, product := range products {
		qty := items[product.ID]

		item, err := entity.NewOrderItem(
			product.ID,
			product.Name,
			product.UnitPrice,
			qty,
			product.UnitOfType,
		)

		if err != nil {
			return err
		}

		if err := order.AddItem(*item); err != nil {
			return err
		}
	}

	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	uc.eventBus.Publish(ctx, utils.Event{
		Name:    "OrderItemAdded",
		Payload: order.ID,
	})

	return nil
}

func (uc *OrderUseCase) RemoveOrderItems(ctx context.Context, orderID uuid.UUID, itemsIDs []uuid.UUID) error {
	if len(itemsIDs) == 0 {
		return ErrEmptyItems
	}

	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	for _, itemID := range itemsIDs {
		if err := order.RemoveItem(itemID); err != nil {
			return err
		}
	}

	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	uc.eventBus.Publish(ctx, utils.Event{
		Name:    "OrderItemRemoved",
		Payload: order.ID,
	})

	return nil
}

func (uc *OrderUseCase) PayOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := order.Pay(); err != nil {
		return err
	}

	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	var itemToPayload []events.ItemPayload

	for _, item := range order.Items() {
		itemInfo := events.ItemPayload{
			ID:       item.ID.String(),
			Name:     item.Name,
			Quantity: item.Quantity,
		}

		itemToPayload = append(itemToPayload, itemInfo)
	}

	uc.eventBus.Publish(ctx, utils.Event{
		Name: "OrderPaid",
		Payload: events.OrderPaidPayload{
			CustomerName: order.CustomerName,
			OrderID:      order.ID.String(),
			Items:        itemToPayload,
		},
	})

	return nil
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := uc.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	if err := order.Cancel(); err != nil {
		return err
	}

	if err := uc.orderRepo.Save(ctx, order); err != nil {
		return err
	}

	uc.eventBus.Publish(ctx, utils.Event{
		Name:    "OrderCancelled",
		Payload: order.ID,
	})

	return nil
}

func (uc *OrderUseCase) ListOrdersByStatus(ctx context.Context, status value_objects.OrderStatus, pagination utils.Pagination) (*utils.PaginatedResult[entity.Order], error) {
	orders, err := uc.orderRepo.FindByStatus(ctx, status, pagination)
	if err != nil {
		return nil, err
	}

	if len(orders.Items) == 0 {
		return nil, ErrNoOrderFound
	}

	return orders, nil
}
