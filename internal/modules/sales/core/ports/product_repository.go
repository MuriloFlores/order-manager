package ports

import (
	"context"

	"github.com/google/uuid"
)

type ProductRepository interface {
	FindAllByIDs(ctx context.Context, productIDs []uuid.UUID) ([]ProductDTO, error)
}
