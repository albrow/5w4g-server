package models

type Item struct {
	Name          string  `json:"name" zoom:"index"`
	ImageUrl      string  `json:"imageUrl"`
	Price         float64 `json:"price"`
	Description   string  `json:"description"`
	AmountInStock int     `json:"amountInStock,omitempty"`
	AmountOrdered int     `json:"amountOrdered,omitempty"`
	Identifier
}
