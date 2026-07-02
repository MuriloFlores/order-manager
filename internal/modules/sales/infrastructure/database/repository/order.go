package repository

import (
	"context"
	"order-manager/internal/modules/sales/core/entity"
	"order-manager/internal/modules/sales/core/ports"
	"order-manager/internal/modules/sales/core/value_objects"
	"order-manager/internal/modules/sales/infrastructure/database/models"
	"order-manager/internal/shared/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewPostgresOrderRepo(db *gorm.DB) ports.OrderRepository {
	return &PostgresOrderRepo{db: db}
}

func (p *PostgresOrderRepo) Save(ctx context.Context, order *entity.Order) error {
	err := p.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(new(models.OrderToModel(order))).Error
	return err
}

func (p *PostgresOrderRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	var model models.OrderModel

	err := p.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id.String()).
		First(&model).Error

	if err != nil {
		return nil, err
	}

	var items []entity.OrderItem

	for _, item := range model.Items {
		itemID, err := uuid.Parse(item.ID)
		if err != nil {
			return nil, err
		}

		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return nil, err
		}

		unit, err := value_objects.NewUnitOfType(item.UnitOfType)
		if err != nil {
			return nil, err
		}

		items = append(items, entity.RestoreOrderItem(
			itemID,
			productID,
			item.Name,
			item.UnitPrice,
			item.Quantity,
			unit,
		))
	}

	orderID, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, err
	}

	status, err := value_objects.NewOrderStatus(model.Status)
	if err != nil {
		return nil, err
	}

	return entity.RestoreOrder(
		orderID,
		model.CustomerName,
		status,
		model.TotalValue,
		model.CreatedAt,
		items,
	), nil
}

func (p *PostgresOrderRepo) FindByStatus(ctx context.Context, status value_objects.OrderStatus, pagination utils.Pagination) (*utils.PaginatedResult[entity.Order], error) {
	var modelsList []models.OrderModel
	var totalCount int64

	p.db.WithContext(ctx).Model(&models.OrderModel{}).Where("status = ?", status).Count(&totalCount)

	if totalCount == 0 {
		return utils.NewPaginatedResult([]entity.Order{}, 0, pagination), nil
	}

	err := p.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", status.String()).
		Offset(pagination.GetOffset()).
		Limit(pagination.GetLimit()).
		Find(&modelsList).Error

	if err != nil {
		return nil, err
	}

	var orders []entity.Order

	for _, model := range modelsList {
		var items []entity.OrderItem

		for _, item := range model.Items {
			itemID, err := uuid.Parse(item.ID)
			if err != nil {
				return nil, err
			}

			productID, err := uuid.Parse(item.ProductID)
			if err != nil {
				return nil, err
			}

			unit, err := value_objects.NewUnitOfType(item.UnitOfType)
			if err != nil {
				return nil, err
			}

			item := entity.RestoreOrderItem(
				itemID,
				productID,
				item.Name,
				item.UnitPrice,
				item.Quantity,
				unit,
			)

			items = append(items, item)
		}

		orderID, err := uuid.Parse(model.ID)
		if err != nil {
			return nil, err
		}

		status, err := value_objects.NewOrderStatus(model.Status)
		if err != nil {
			return nil, err
		}

		orderEntity := entity.RestoreOrder(
			orderID,
			model.CustomerName,
			status,
			model.TotalValue,
			model.CreatedAt,
			items,
		)

		orders = append(orders, *orderEntity)
	}

	return utils.NewPaginatedResult(orders, totalCount, pagination), nil
}
