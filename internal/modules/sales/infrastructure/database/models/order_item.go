package models

import "order-manager/internal/modules/sales/core/entity"

type OrderItemModel struct {
	ID         string  `gorm:"type:uuid;primaryKey"`
	OrderID    string  `gorm:"type:uuid;index;not null"`
	ProductID  string  `gorm:"type:uuid;not null"`
	Name       string  `gorm:"type:varchar(255);not null"`
	UnitPrice  float64 `gorm:"type:numeric(10,2)"`
	Quantity   float64 `gorm:"type:numeric(10,2)"`
	UnitOfType string  `gorm:"type:varchar(50)"`
}

func (OrderItemModel) TableName() string {
	return "orders_items"
}

func OrderItemToModel(orderItem entity.OrderItem) OrderItemModel {
	return OrderItemModel{
		ID:         orderItem.ID.String(),
		ProductID:  orderItem.ProductID.String(),
		Name:       orderItem.Name,
		UnitPrice:  orderItem.UnitPrice,
		Quantity:   orderItem.Quantity,
		UnitOfType: orderItem.UnitOfType.String(),
	}
}
