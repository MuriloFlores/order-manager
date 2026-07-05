package ports

import (
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

type ProductDTO struct {
	ID         uuid.UUID
	Name       string
	UnitPrice  float64
	UnitOfType value_objects.UnitOfType
}
