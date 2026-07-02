package catalog

import (
	"errors"

	"github.com/google/uuid"
)

type CatalogDTO struct {
	ID    uuid.UUID
	Name  string
	Price float64
}

type CatalogService interface {
	GetProductsData(ids []uuid.UUID) ([]CatalogDTO, error)
}

type FakeCatalogService struct {
	fakeDatabase map[uuid.UUID]CatalogDTO
}

func NewFakeCatalogService() *FakeCatalogService {
	db := make(map[uuid.UUID]CatalogDTO)

	return &FakeCatalogService{
		fakeDatabase: db,
	}
}

func (s *FakeCatalogService) GetProductsData(ids []uuid.UUID) ([]CatalogDTO, error) {
	if len(ids) == 0 {
		return nil, errors.New("no ids provided to catalog")
	}

	var results []CatalogDTO
	for _, id := range ids {
		results = append(results, CatalogDTO{
			ID:    id,
			Name:  "Produto Dinâmico do Catálogo Falso",
			Price: 25.50,
		})
	}

	return results, nil
}
