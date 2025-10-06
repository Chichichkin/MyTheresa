package catalog

import "github.com/mytheresa/go-hiring-challenge/models"

type Response struct {
	Products []Product `json:"products"`
	// From  models POV products != variants. But from clients perspective  each variant is a different product
	// Without any additional clarification I assume that products_available means total number of products * variants
	ProductsAvailable int              `json:"products_available"`
	Variants          []models.Variant `json:"variants,omitempty"`
}

type Product struct {
	Code     string  `json:"code"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}
