package models

type Item struct {
	Name          string  `json:"name" zoom:"index"`
	ImageUrl      string  `json:"imageUrl"` // A public-facing url which can be used to get the image file
	ImageOrigPath string  `json:"-"`        // The original path of the image file, which is used internally to store it on s3
	Price         float64 `json:"price"`
	Description   string  `json:"description"`
	AmountInStock int     `json:"amountInStock,omitempty"`
	AmountOrdered int     `json:"amountOrdered,omitempty"`
	Identifier
}
