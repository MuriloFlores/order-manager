package adapters

import (
	"context"
	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/core/ports"
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

type CatalogGateway struct {
	catalogAPI catalog.CatalogService
}

func NewCatalogGateway(api catalog.CatalogService) *CatalogGateway {
	return &CatalogGateway{catalogAPI: api}
}

func (g *CatalogGateway) FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]ports.ProductDTO, error) {
	catalogData, err := g.catalogAPI.GetProductsData(ctx, ids)
	if err != nil {
		return nil, err
	}

	var products []ports.ProductDTO
	for _, item := range catalogData {
		unit, err := value_objects.NewUnitOfType(item.UnitOfType)
		if err != nil {
			return nil, err
		}

		prod := ports.ProductDTO{
			ID:         item.ID,
			Name:       item.Name,
			UnitPrice:  item.Price,
			UnitOfType: unit,
		}

		products = append(products, prod)
	}

	return products, nil
}
