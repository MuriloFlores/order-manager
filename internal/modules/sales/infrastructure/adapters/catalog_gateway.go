package adapters

import (
	"context"
	"order-manager/internal/modules/catalog"
	"order-manager/internal/modules/sales/core/entity"
	"order-manager/internal/modules/sales/core/value_objects"

	"github.com/google/uuid"
)

type CatalogGateway struct {
	catalogAPI catalog.CatalogService
}

func NewCatalogGateway(api catalog.CatalogService) *CatalogGateway {
	return &CatalogGateway{catalogAPI: api}
}

func (g *CatalogGateway) FindAllByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error) {
	catalogData, err := g.catalogAPI.GetProductsData(ids)
	if err != nil {
		return nil, err
	}

	var products []entity.Product
	for _, item := range catalogData {
		products = append(products, entity.Product{
			ID:         item.ID,
			Name:       item.Name,
			UnitPrice:  item.Price,
			UnitOfType: value_objects.Unit, 
		})
	}

	return products, nil
}
