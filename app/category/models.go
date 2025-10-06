package category

type Response struct {
	Categories []Category `json:"categories"`
}

type Category struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type CreateRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
