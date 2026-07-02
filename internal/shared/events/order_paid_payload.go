package events

type OrderPaidPayload struct {
	CustomerName string        `json:"customer_name"`
	OrderID      string        `json:"order_id"`
	Items        []ItemPayload `json:"items"`
}

type ItemPayload struct {
	ID       string  `json:"item_id"`
	Name     string  `json:"item_name"`
	Quantity float64 `json:"quantity"`
}
