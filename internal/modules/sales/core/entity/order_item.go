package entity

import (
	"errors"
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

var (
	ErrInvalidItemQuantity = errors.New("invalid item quantity")
	ErrInvalidProductID    = errors.New("invalid order id")
)

type OrderItem struct {
	ID         uuid.UUID
	ProductID  uuid.UUID
	Name       string
	UnitPrice  float64
	Quantity   float64
	UnitOfType value_objects.UnitOfType
}

func NewOrderItem(productID uuid.UUID, name string, unitPrice float64, quantity float64, unitOfType value_objects.UnitOfType) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, ErrInvalidItemQuantity
	}

	if productID == uuid.Nil {
		return nil, ErrInvalidProductID
	}

	return &OrderItem{
		ID:         uuid.New(),
		ProductID:  productID,
		Name:       name,
		UnitPrice:  unitPrice,
		Quantity:   quantity,
		UnitOfType: unitOfType,
	}, nil
}

func RestoreOrderItem(id, productID uuid.UUID, name string, unitPrice, quantity float64, unitOfType value_objects.UnitOfType) OrderItem {
	return OrderItem{
		ID:         id,
		ProductID:  productID,
		Name:       name,
		UnitPrice:  unitPrice,
		Quantity:   quantity,
		UnitOfType: unitOfType,
	}
}

func (i *OrderItem) SubTotal() float64 {
	return i.UnitPrice * i.Quantity
}
