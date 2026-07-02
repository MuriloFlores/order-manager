package ports

import (
	"context"
	"order-manager/internal/modules/sales/core/entity"
	"order-manager/internal/modules/sales/core/value_objects"
	"order-manager/internal/shared/utils"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Save(ctx context.Context, order *entity.Order) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	FindByStatus(ctx context.Context, status value_objects.OrderStatus, pagination utils.Pagination) (*utils.PaginatedResult[entity.Order], error)
}
