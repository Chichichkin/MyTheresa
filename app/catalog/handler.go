package catalog

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/app/repos/products"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
)

type Handler struct {
	repo products.Repository
}

func NewCatalogHandler(r products.Repository) *Handler {
	return &Handler{
		repo: r,
	}
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	filters := validateProductFilters(
		r.URL.Query().Get("offset"),
		r.URL.Query().Get("limit"),
		r.URL.Query().Get("priceLessThan"),
		r.URL.Query().Get("category"),
	)

	dbProducts, err := h.repo.List(r.Context(), filters)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := prepareResponse(dbProducts, false)

	api.OKResponse(w, response)
}

func (h *Handler) HandleGetSpecific(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	// Validate product code format before making database call
	if !validateProductCode(code) {
		api.ErrorResponse(
			w,
			http.StatusBadRequest,
			"Invalid product code format. Expected format: PROD followed by 3 digits (e.g., PROD001)",
		)
		return
	}

	dbProducts, err := h.repo.GetByCode(r.Context(), code)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := prepareResponse([]models.Product{dbProducts}, true)

	api.OKResponse(w, response)
}

func prepareResponse(dbProducts []models.Product, includeVariants bool) Response {
	respProducts := make([]Product, len(dbProducts))
	uniqueProductCount := 0
	for i, p := range dbProducts {
		respProducts[i] = Product{
			Code:     p.Code,
			Price:    p.Price.InexactFloat64(),
			Category: p.Category.Name,
		}
		if len(p.Variants) > 0 {
			uniqueProductCount += len(p.Variants)
		}
	}

	resp := Response{Products: respProducts, ProductsAvailable: uniqueProductCount}

	// Including variants only for the first product as per the requirement
	// If needed for all products, this logic can be adjusted

	if includeVariants && len(dbProducts) > 0 {
		resp.Variants = make([]models.Variant, 0, len(dbProducts[0].Variants))
		for _, variant := range dbProducts[0].Variants {
			if variant.Price == decimal.Zero {
				variant.Price = dbProducts[0].Price
			}
			resp.Variants = append(resp.Variants, variant)
		}
	}
	return resp
}

func validateProductFilters(offset, limit, priceLimit, category string) products.SearchFilters {
	filters := products.SearchFilters{
		Offset:   0,
		Limit:    10,
		Category: "",
	}

	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err == nil && o >= 0 {
			filters.Offset = o
		}
	}
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err == nil && l > 0 && l <= 100 {
			filters.Limit = l
		}
	}

	if priceLimit != "" {
		p, err := decimal.NewFromString(priceLimit)
		if err == nil && p.GreaterThan(decimal.Zero) {
			filters.PriceLessThan = &p
		}
	}

	if category != "" {
		filters.Category = category
	}

	return filters
}

// validateProductCode validates if the product code follows the expected format
// Valid format: PROD followed by 3 digits (e.g., PROD001, PROD123)
func validateProductCode(code string) bool {
	if code == "" {
		return false
	}

	pattern := `^PROD\d{3}$`
	matched, err := regexp.MatchString(pattern, code)
	if err != nil {
		return false
	}

	return matched
}
