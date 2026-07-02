package models

import (
	"order-manager/internal/modules/sales/core/entity"
	"time"
)

type OrderModel struct {
	ID           string  `gorm:"type:uuid;primaryKey"`
	CustomerName string  `gorm:"type:varchar(255);not null"`
	Status       string  `gorm:"type:varchar(50);not null"`
	TotalValue   float64 `gorm:"type:numeric(10,2);not null"`
	CreatedAt    time.Time

	Items []OrderItemModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string {
	return "orders"
}

func OrderToModel(order *entity.Order) OrderModel {
	var itemsModel []OrderItemModel

	for _, item := range order.Items() {
		itemModel := OrderItemToModel(item)

		itemModel.OrderID = order.ID.String()

		itemsModel = append(itemsModel, itemModel)
	}

	return OrderModel{
		ID:           order.ID.String(),
		CustomerName: order.CustomerName,
		Items:        itemsModel,
		Status:       order.Status().String(),
		TotalValue:   order.TotalValue(),
		CreatedAt:    order.CreatedAt(),
	}
}
