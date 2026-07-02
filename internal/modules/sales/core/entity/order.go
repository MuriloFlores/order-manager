package entity

import (
	"errors"
	"order-manager/internal/modules/sales/core/value_objects"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderAlreadyPaid      = errors.New("order already paid")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")
	ErrOrderAlreadyProcessed = errors.New("order already processed")
	ErrItemNotFound          = errors.New("item not found")
)

type Order struct {
	ID           uuid.UUID
	CustomerName string
	items        []OrderItem
	status       value_objects.OrderStatus
	totalValue   float64
	createdAt    time.Time
}

func NewOrder(customerName string) *Order {
	return &Order{
		ID:           uuid.New(),
		CustomerName: customerName,
		items:        []OrderItem{},
		status:       value_objects.StatusPending,
		createdAt:    time.Now(),
	}
}

func RestoreOrder(id uuid.UUID, customerName string, status value_objects.OrderStatus, total float64, created time.Time, items []OrderItem) *Order {
	return &Order{
		ID:           id,
		CustomerName: customerName,
		status:       status,
		totalValue:   total,
		createdAt:    created,
		items:        items,
	}
}

func (o *Order) TotalValue() float64 {
	return o.totalValue
}

func (o *Order) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Order) Status() value_objects.OrderStatus {
	return o.status
}

func (o *Order) Items() []OrderItem {
	return o.items
}

func (o *Order) CalculateTotalValue() {
	var totalValue float64

	for _, item := range o.items {
		totalValue += item.SubTotal()
	}

	o.totalValue = totalValue
}

func (o *Order) AddItem(item OrderItem) error {
	if o.status != value_objects.StatusPending {
		return ErrOrderAlreadyProcessed
	}

	o.items = append(o.items, item)
	o.CalculateTotalValue()

	return nil
}

func (o *Order) RemoveItem(itemID uuid.UUID) error {
	if o.status != value_objects.StatusPending {
		return ErrOrderAlreadyProcessed
	}

	indexToRemove := -1
	for i, item := range o.items {
		if item.ID == itemID {
			indexToRemove = i
			break
		}
	}

	if indexToRemove == -1 {
		return ErrItemNotFound
	}

	o.items = append(o.items[:indexToRemove], o.items[indexToRemove+1:]...)

	o.CalculateTotalValue()

	return nil
}

func (o *Order) Pay() error {
	if o.status == value_objects.StatusPaid {
		return ErrOrderAlreadyPaid
	}

	if o.status == value_objects.StatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	o.status = value_objects.StatusPaid

	return nil
}

func (o *Order) Cancel() error {
	if o.status == value_objects.StatusPaid {
		return ErrOrderAlreadyPaid
	}

	if o.status == value_objects.StatusCancelled {
		return ErrOrderAlreadyCancelled
	}

	o.status = value_objects.StatusCancelled

	return nil
}
