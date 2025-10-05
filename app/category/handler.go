package category

import (
	"encoding/json"
	"net/http"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/repos/category"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type Handler struct {
	repo category.Repository
}

func NewCategoryHandler(r category.Repository) *Handler {
	return &Handler{
		repo: r,
	}
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.ListAll(r.Context())
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := prepareResponse(categories)
	api.OKResponse(w, response)
}

func (h *Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Code == "" || req.Name == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Code and name are required")
		return
	}

	newCategory := models.Category{
		Code: req.Code,
		Name: req.Name,
	}

	if err := h.repo.Create(r.Context(), newCategory); err != nil {
		if err.Error() == "category code already exists" {
			api.ErrorResponse(w, http.StatusConflict, err.Error())
			return
		}
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.OKResponse(w, map[string]string{"message": "Category created successfully"})
}

func prepareResponse(categories []models.Category) Response {
	categoryResponses := make([]Category, len(categories))
	for i, cat := range categories {
		categoryResponses[i] = Category{
			Code: cat.Code,
			Name: cat.Name,
		}
	}
	return Response{Categories: categoryResponses}
}
