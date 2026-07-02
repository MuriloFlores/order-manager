package value_objects

import (
	"errors"
	"strings"
)

type UnitOfType string

var (
	ErrInvalidUnitOfType = errors.New("invalid unit of type")
)

const (
	Grams      UnitOfType = "G"
	Kilos      UnitOfType = "KG"
	Unit       UnitOfType = "UND"
	Milliliter UnitOfType = "ML"
)

func NewUnitOfType(unitOfType string) (UnitOfType, error) {
	normalizedUnit := UnitOfType(strings.TrimSpace(strings.ToUpper(unitOfType)))

	switch normalizedUnit {
	case Grams, Kilos, Unit, Milliliter:
		return normalizedUnit, nil
	default:
		return "", ErrInvalidUnitOfType
	}
}

func (u UnitOfType) String() string {
	return string(u)
}

func (u UnitOfType) Equals(other UnitOfType) bool {
	return u.String() == other.String()
}
