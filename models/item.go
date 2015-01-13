package models

type Item struct {
	Name          string  `json:"name" zoom:"index"`
	ImageUrl      string  `json:"imageUrl"` // A public-facing url which can be used to get the image file
	ImageS3Path   string  `json:"-"`        // The path of the image file stored on s3. Only used internally
	Price         float64 `json:"price"`
	Description   string  `json:"description"`
	AmountInStock int     `json:"amountInStock,omitempty"`
	AmountOrdered int     `json:"amountOrdered,omitempty"`
	Identifier
}
