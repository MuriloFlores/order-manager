package value_objects

import (
	"errors"
	"strings"
)

type OrderStatus string

var (
	ErrInvalidOrderStatus = errors.New("invalid order status")
)

const (
	StatusPending   OrderStatus = "PENDING"
	StatusCancelled OrderStatus = "CANCELLED"
	StatusPaid      OrderStatus = "PAID"
)

func NewOrderStatus(status string) (OrderStatus, error) {
	normalizedStatus := OrderStatus(strings.TrimSpace(strings.ToUpper(status)))

	switch normalizedStatus {
	case StatusPaid, StatusCancelled, StatusPending:
		return normalizedStatus, nil
	default:
		return "", ErrInvalidOrderStatus
	}
}

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) Equals(other OrderStatus) bool {
	return s.String() == other.String()
}
