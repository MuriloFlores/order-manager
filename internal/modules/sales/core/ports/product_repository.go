package ports

import (
	"context"
	"order-manager/internal/modules/sales/core/entity"

	"github.com/google/uuid"
)

type ProductRepository interface {
	FindAllByIDs(ctx context.Context, productIDs []uuid.UUID) ([]entity.Product, error)
}
