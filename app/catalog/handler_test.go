package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/app/repos/products"
	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type MockProductRepo struct {
	ListAllFunc       func(ctx context.Context) ([]models.Product, error)
	ListFunc          func(ctx context.Context, filters products.SearchFilters) ([]models.Product, error)
	GetByIDFunc       func(ctx context.Context, id string) (models.Product, error)
	GetByCodeFunc     func(ctx context.Context, code string) (models.Product, error)
	GetByCategoryFunc func(ctx context.Context, category string) ([]models.Product, error)
}

func (m *MockProductRepo) ListAll(ctx context.Context) ([]models.Product, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockProductRepo) List(ctx context.Context, filters products.SearchFilters) ([]models.Product, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, filters)
	}
	return nil, nil
}

func (m *MockProductRepo) GetByID(ctx context.Context, id string) (models.Product, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return models.Product{}, nil
}

func (m *MockProductRepo) GetByCode(ctx context.Context, code string) (models.Product, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return models.Product{}, nil
}

func (m *MockProductRepo) GetByCategory(ctx context.Context, category string) ([]models.Product, error) {
	if m.GetByCategoryFunc != nil {
		return m.GetByCategoryFunc(ctx, category)
	}
	return nil, nil
}

func TestHandler_HandleGet(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockProducts   []models.Product
		mockError      error
		expectedStatus int
		expectedBody   Response
	}{
		{
			name:        "successful get products with default filters",
			queryParams: "",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
				{
					ID:    2,
					Code:  "PROD002",
					Price: decimal.NewFromFloat(200.75),
					Category: models.Category{
						ID:   2,
						Code: "shoes",
						Name: "Shoes",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Products: []Product{
					{Code: "PROD001", Price: 100.50, Category: "Clothing"},
					{Code: "PROD002", Price: 200.75, Category: "Shoes"},
				},
				ProductsAvailable: 0, // No variants in this test case
			},
		},
		{
			name:        "successful get products with filters",
			queryParams: "?offset=10&limit=5&priceLessThan=150&category=clothing",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Products: []Product{
					{Code: "PROD001", Price: 100.50, Category: "Clothing"},
				},
				ProductsAvailable: 0, // No variants in this test case
			},
		},
		{
			name:           "empty products list",
			queryParams:    "",
			mockProducts:   []models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Products:          []Product{},
				ProductsAvailable: 0,
			},
		},
		{
			name:           "database error",
			queryParams:    "",
			mockProducts:   nil,
			mockError:      errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   Response{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProductRepo{
				ListFunc: func(ctx context.Context, filters products.SearchFilters) ([]models.Product, error) {
					return tt.mockProducts, tt.mockError
				},
			}

			handler := NewCatalogHandler(mockRepo)

			req := httptest.NewRequest("GET", "/catalog"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleGet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestHandler_HandleGetSpecific(t *testing.T) {
	tests := []struct {
		name           string
		productCode    string
		mockProduct    models.Product
		mockError      error
		expectedStatus int
		expectedBody   Response
	}{
		{
			name:        "successful get specific product with variants",
			productCode: "PROD001",
			mockProduct: models.Product{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.NewFromFloat(100.50),
				Category: models.Category{
					ID:   1,
					Code: "clothing",
					Name: "Clothing",
				},
				Variants: []models.Variant{
					{
						ID:        1,
						Name:      "Medium",
						SKU:       "PROD001-M",
						Price:     decimal.Zero, // Should inherit from product
						ProductID: 1,
					},
					{
						ID:        2,
						Name:      "Large",
						SKU:       "PROD001-L",
						Price:     decimal.NewFromFloat(120), // Has specific price
						ProductID: 1,
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Products: []Product{
					{Code: "PROD001", Price: 100.50, Category: "Clothing"},
				},
				ProductsAvailable: 2, // Count of variants
				Variants: []models.Variant{
					{
						ID:        1,
						Name:      "Medium",
						SKU:       "PROD001-M",
						Price:     decimal.NewFromFloat(100.50), // Inherited from product
						ProductID: 1,
					},
					{
						ID:        2,
						Name:      "Large",
						SKU:       "PROD001-L",
						Price:     decimal.NewFromFloat(120), // Specific price
						ProductID: 1,
					},
				},
			},
		},
		{
			name:        "successful get specific product without variants",
			productCode: "PROD002",
			mockProduct: models.Product{
				ID:    2,
				Code:  "PROD002",
				Price: decimal.NewFromFloat(200.75),
				Category: models.Category{
					ID:   2,
					Code: "shoes",
					Name: "Shoes",
				},
				Variants: []models.Variant{},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Products: []Product{
					{Code: "PROD002", Price: 200.75, Category: "Shoes"},
				},
				ProductsAvailable: 0, // No variants
				Variants:          nil,
			},
		},
		{
			name:           "product not found",
			productCode:    "PROD999",
			mockProduct:    models.Product{},
			mockError:      errors.New("product not found"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   Response{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProductRepo{
				GetByCodeFunc: func(ctx context.Context, code string) (models.Product, error) {
					return tt.mockProduct, tt.mockError
				},
			}

			handler := NewCatalogHandler(mockRepo)

			req := httptest.NewRequest("GET", "/catalog/"+tt.productCode, nil)
			req.SetPathValue("code", tt.productCode)
			w := httptest.NewRecorder()

			handler.HandleGetSpecific(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedBody.Products), len(response.Products))
				assert.Equal(t, tt.expectedBody.ProductsAvailable, response.ProductsAvailable)
				for i, product := range response.Products {
					assert.Contains(t, tt.expectedBody.Products[i].Code, product.Code)
				}
			}
		})
	}
}

func TestValidateProductFilters(t *testing.T) {
	tests := []struct {
		name           string
		offset         string
		limit          string
		priceLimit     string
		category       string
		expectedResult products.SearchFilters
	}{
		{
			name:       "default values",
			offset:     "",
			limit:      "",
			priceLimit: "",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        0,
				Limit:         10,
				Category:      "",
				PriceLessThan: nil,
			},
		},
		{
			name:       "valid offset and limit",
			offset:     "20",
			limit:      "25",
			priceLimit: "",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        20,
				Limit:         25,
				Category:      "",
				PriceLessThan: nil,
			},
		},
		{
			name:       "invalid offset (negative)",
			offset:     "-5",
			limit:      "15",
			priceLimit: "",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        0, // Should remain default
				Limit:         15,
				Category:      "",
				PriceLessThan: nil,
			},
		},
		{
			name:       "invalid limit (too high)",
			offset:     "10",
			limit:      "150",
			priceLimit: "",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        10,
				Limit:         10, // Should remain default
				Category:      "",
				PriceLessThan: nil,
			},
		},
		{
			name:       "invalid limit (zero)",
			offset:     "10",
			limit:      "0",
			priceLimit: "",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        10,
				Limit:         10, // Should remain default
				Category:      "",
				PriceLessThan: nil,
			},
		},
		{
			name:       "valid price limit",
			offset:     "0",
			limit:      "10",
			priceLimit: "99.99",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:   0,
				Limit:    10,
				Category: "",
				PriceLessThan: func() *decimal.Decimal {
					d := decimal.NewFromFloat(99.99)
					return &d
				}(),
			},
		},
		{
			name:       "invalid price limit (negative)",
			offset:     "0",
			limit:      "10",
			priceLimit: "-50.00",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        0,
				Limit:         10,
				Category:      "",
				PriceLessThan: nil, // Should be nil for negative price
			},
		},
		{
			name:       "invalid price limit (non-numeric)",
			offset:     "0",
			limit:      "10",
			priceLimit: "invalid",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        0,
				Limit:         10,
				Category:      "",
				PriceLessThan: nil, // Should be nil for invalid price
			},
		},
		{
			name:       "valid category",
			offset:     "0",
			limit:      "10",
			priceLimit: "",
			category:   "clothing",
			expectedResult: products.SearchFilters{
				Offset:        0,
				Limit:         10,
				Category:      "clothing",
				PriceLessThan: nil,
			},
		},
		{
			name:       "all parameters valid",
			offset:     "15",
			limit:      "20",
			priceLimit: "199.99",
			category:   "shoes",
			expectedResult: products.SearchFilters{
				Offset:   15,
				Limit:    20,
				Category: "shoes",
				PriceLessThan: func() *decimal.Decimal {
					d := decimal.NewFromFloat(199.99)
					return &d
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateProductFilters(tt.offset, tt.limit, tt.priceLimit, tt.category)
			assert.Equal(t, tt.expectedResult.Category, result.Category)
			assert.Equal(t, tt.expectedResult.Offset, result.Offset)
			assert.Equal(t, tt.expectedResult.Limit, result.Limit)
			if tt.expectedResult.PriceLessThan != nil {
				assert.NotNil(t, result.PriceLessThan)
				assert.True(t, tt.expectedResult.PriceLessThan.Equal(*result.PriceLessThan))
			}
		})
	}
}

func TestPrepareResponse(t *testing.T) {
	tests := []struct {
		name             string
		dbProducts       []models.Product
		includeVariants  bool
		expectedResponse Response
	}{
		{
			name: "products without variants",
			dbProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
				{
					ID:    2,
					Code:  "PROD002",
					Price: decimal.NewFromFloat(200.75),
					Category: models.Category{
						ID:   2,
						Code: "shoes",
						Name: "Shoes",
					},
				},
			},
			includeVariants: false,
			expectedResponse: Response{
				Products: []Product{
					{Code: "PROD001", Price: 100.50, Category: "Clothing"},
					{Code: "PROD002", Price: 200.75, Category: "Shoes"},
				},
				ProductsAvailable: 0, // No variants in this test case
			},
		},
		{
			name: "products with variants (inherit price)",
			dbProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
					Variants: []models.Variant{
						{
							ID:        1,
							Name:      "Medium",
							SKU:       "PROD001-M",
							Price:     decimal.Zero, // Should inherit from product
							ProductID: 1,
						},
						{
							ID:        2,
							Name:      "Large",
							SKU:       "PROD001-L",
							Price:     decimal.NewFromFloat(120.00), // Has specific price
							ProductID: 1,
						},
					},
				},
			},
			includeVariants: true,
			expectedResponse: Response{
				Products: []Product{
					{Code: "PROD001", Price: 100.50, Category: "Clothing"},
				},
				ProductsAvailable: 2, // Count of variants
				Variants: []models.Variant{
					{
						ID:        1,
						Name:      "Medium",
						SKU:       "PROD001-M",
						Price:     decimal.NewFromFloat(100.50), // Inherited from product
						ProductID: 1,
					},
					{
						ID:        2,
						Name:      "Large",
						SKU:       "PROD001-L",
						Price:     decimal.NewFromFloat(120), // Specific price
						ProductID: 1,
					},
				},
			},
		},
		{
			name:             "empty products list",
			dbProducts:       []models.Product{},
			includeVariants:  false,
			expectedResponse: Response{Products: []Product{}, ProductsAvailable: 0},
		},
		{
			name: "include variants but no variants",
			dbProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
					Variants: []models.Variant{},
				},
			},
			includeVariants: true,
			expectedResponse: Response{
				Products: []Product{
					{Code: "PROD001", Price: 100.50, Category: "Clothing"},
				},
				ProductsAvailable: 0, // No variants
				Variants:          []models.Variant{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prepareResponse(tt.dbProducts, tt.includeVariants)
			assert.Equal(t, len(tt.expectedResponse.Products), len(result.Products))
			assert.Equal(t, tt.expectedResponse.ProductsAvailable, result.ProductsAvailable)
			for i, product := range result.Products {
				assert.Contains(t, tt.expectedResponse.Products[i].Code, product.Code)
			}
		})
	}
}

func TestHandler_HandleGet_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockProducts   []models.Product
		mockError      error
		expectedStatus int
		description    string
	}{
		{
			name:        "negative offset",
			queryParams: "?offset=-5",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should default to offset=0 when negative value provided",
		},
		{
			name:        "zero limit",
			queryParams: "?limit=0",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should default to limit=10 when zero provided",
		},
		{
			name:        "limit over maximum",
			queryParams: "?limit=150",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should cap limit at 100 when value exceeds maximum",
		},
		{
			name:        "negative price filter",
			queryParams: "?priceLessThan=-50",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should ignore negative price filter",
		},
		{
			name:        "invalid price format",
			queryParams: "?priceLessThan=not-a-number",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should ignore invalid price format",
		},
		{
			name:           "very large offset",
			queryParams:    "?offset=999999",
			mockProducts:   []models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should handle very large offset values",
		},
		{
			name:        "empty category filter",
			queryParams: "?category=",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should handle empty category filter",
		},
		{
			name:           "special characters in category",
			queryParams:    "?category=clothing%20%26%20shoes",
			mockProducts:   []models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should handle URL-encoded special characters in category",
		},
		{
			name:        "multiple same parameters",
			queryParams: "?limit=5&limit=10&offset=0&offset=5",
			mockProducts: []models.Product{
				{
					ID:    1,
					Code:  "PROD001",
					Price: decimal.NewFromFloat(100.50),
					Category: models.Category{
						ID:   1,
						Code: "clothing",
						Name: "Clothing",
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should handle duplicate query parameters (uses first value)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProductRepo{
				ListFunc: func(ctx context.Context, filters products.SearchFilters) ([]models.Product, error) {
					return tt.mockProducts, tt.mockError
				},
			}

			handler := NewCatalogHandler(mockRepo)

			req := httptest.NewRequest("GET", "/catalog"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.HandleGet(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Response should be valid JSON")
				assert.NotNil(t, response.Products, "Products should not be nil")
			}
		})
	}
}

func TestHandler_HandleGetSpecific_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		productCode    string
		mockProduct    models.Product
		mockError      error
		expectedStatus int
		description    string
	}{
		{
			name:           "empty product code",
			productCode:    "",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty product code with validation error",
		},
		{
			name:           "product code with special characters",
			productCode:    "PROD-001",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject product code with special characters",
		},
		{
			name:           "unicode product code",
			productCode:    "ПРОД001",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject unicode characters in product code",
		},
		{
			name:           "invalid format - too short",
			productCode:    "PROD01",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject product code that's too short",
		},
		{
			name:           "invalid format - too long",
			productCode:    "PROD1234",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject product code that's too long",
		},
		{
			name:           "invalid format - wrong prefix",
			productCode:    "ITEM001",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject product code with wrong prefix",
		},
		{
			name:           "invalid format - non-numeric suffix",
			productCode:    "PRODABC",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject product code with non-numeric suffix",
		},
		{
			name:           "valid product code format",
			productCode:    "PROD001",
			mockProduct:    models.Product{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should accept valid product code format",
		},
		{
			name:        "product with zero price",
			productCode: "PROD001",
			mockProduct: models.Product{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.Zero,
				Category: models.Category{
					ID:   1,
					Code: "clothing",
					Name: "Clothing",
				},
				Variants: []models.Variant{
					{
						ID:        1,
						Name:      "Medium",
						SKU:       "PROD001-M",
						Price:     decimal.Zero,
						ProductID: 1,
					},
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should handle products with zero price",
		},
		{
			name:        "product with very high price",
			productCode: "PROD001",
			mockProduct: models.Product{
				ID:    1,
				Code:  "PROD001",
				Price: decimal.NewFromFloat(999999.99),
				Category: models.Category{
					ID:   1,
					Code: "clothing",
					Name: "Clothing",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			description:    "Should handle products with very high prices",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockProductRepo{
				GetByCodeFunc: func(ctx context.Context, code string) (models.Product, error) {
					return tt.mockProduct, tt.mockError
				},
			}

			handler := NewCatalogHandler(mockRepo)

			req := httptest.NewRequest("GET", "/catalog/"+tt.productCode, nil)
			req.SetPathValue("code", tt.productCode)
			w := httptest.NewRecorder()

			handler.HandleGetSpecific(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			if tt.expectedStatus == http.StatusOK {
				var response Response
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Response should be valid JSON")
			}
		})
	}
}

func TestValidateProductFilters_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		offset         string
		limit          string
		priceLimit     string
		category       string
		expectedResult products.SearchFilters
		description    string
	}{
		{
			name:       "all parameters with extreme values",
			offset:     "999999",
			limit:      "999999",
			priceLimit: "999999.99",
			category:   "a",
			expectedResult: products.SearchFilters{
				Offset:   999999,
				Limit:    10, // Will be dropped to default one
				Category: "a",
				PriceLessThan: func() *decimal.Decimal {
					d := decimal.NewFromFloat(999999.99)
					return &d
				}(),
			},
			description: "Should handle extreme values and cap limit at 100",
		},
		{
			name:       "negative values everywhere",
			offset:     "-999",
			limit:      "-999",
			priceLimit: "-999.99",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        0,  // Should default to 0
				Limit:         10, // Should default to 10
				Category:      "",
				PriceLessThan: nil, // Should be nil for negative price
			},
			description: "Should handle negative values and use defaults",
		},
		{
			name:       "zero values",
			offset:     "0",
			limit:      "0",
			priceLimit: "0",
			category:   "",
			expectedResult: products.SearchFilters{
				Offset:        0,
				Limit:         10, // Should default to 10 (minimum is 1)
				Category:      "",
				PriceLessThan: nil, // Should be nil for zero price
			},
			description: "Should handle zero values appropriately",
		},
		{
			name:       "boundary limit values",
			offset:     "0",
			limit:      "1",    // Minimum valid limit
			priceLimit: "0.01", // Minimum positive price
			category:   "a",
			expectedResult: products.SearchFilters{
				Offset:   0,
				Limit:    1,
				Category: "a",
				PriceLessThan: func() *decimal.Decimal {
					d := decimal.NewFromFloat(0.01)
					return &d
				}(),
			},
			description: "Should handle minimum valid values",
		},
		{
			name:       "maximum valid limit",
			offset:     "0",
			limit:      "100", // Maximum valid limit
			priceLimit: "999999.99",
			category:   "a",
			expectedResult: products.SearchFilters{
				Offset:   0,
				Limit:    100,
				Category: "a",
				PriceLessThan: func() *decimal.Decimal {
					d := decimal.NewFromFloat(999999.99)
					return &d
				}(),
			},
			description: "Should handle maximum valid limit",
		},
		{
			name:       "just over maximum limit",
			offset:     "0",
			limit:      "101", // Just over maximum
			priceLimit: "100.00",
			category:   "a",
			expectedResult: products.SearchFilters{
				Offset:   0,
				Limit:    10, // Should default to 10
				Category: "a",
				PriceLessThan: func() *decimal.Decimal {
					d := decimal.NewFromFloat(100.00)
					return &d
				}(),
			},
			description: "Should default to 10 when limit exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateProductFilters(tt.offset, tt.limit, tt.priceLimit, tt.category)
			assert.Equal(t, tt.expectedResult.Category, result.Category, tt.description)
			assert.Equal(t, tt.expectedResult.Offset, result.Offset, tt.description)
			assert.Equal(t, tt.expectedResult.Limit, result.Limit, tt.description)
			if tt.expectedResult.PriceLessThan == nil {
				assert.Nil(t, result.PriceLessThan, tt.description)
			} else {
				assert.NotNil(t, result.PriceLessThan, tt.description)
				assert.True(t, tt.expectedResult.PriceLessThan.Equal(*result.PriceLessThan), tt.description)
			}
		})
	}
}

func TestValidateProductCode(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name:     "valid product code",
			code:     "PROD001",
			expected: true,
		},
		{
			name:     "valid product code with different numbers",
			code:     "PROD123",
			expected: true,
		},
		{
			name:     "valid product code with leading zeros",
			code:     "PROD000",
			expected: true,
		},
		{
			name:     "empty code",
			code:     "",
			expected: false,
		},
		{
			name:     "code too short",
			code:     "PROD01",
			expected: false,
		},
		{
			name:     "code too long",
			code:     "PROD1234",
			expected: false,
		},
		{
			name:     "wrong prefix",
			code:     "ITEM001",
			expected: false,
		},
		{
			name:     "non-numeric suffix",
			code:     "PRODABC",
			expected: false,
		},
		{
			name:     "special characters",
			code:     "PROD-001",
			expected: false,
		},
		{
			name:     "spaces in code",
			code:     "PROD 001",
			expected: false,
		},
		{
			name:     "unicode characters",
			code:     "ПРОД001",
			expected: false,
		},
		{
			name:     "lowercase prefix",
			code:     "prod001",
			expected: false,
		},
		{
			name:     "mixed case",
			code:     "Prod001",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateProductCode(tt.code)
			assert.Equal(t, tt.expected, result, "validateProductCode(%q) should return %v", tt.code, tt.expected)
		})
	}
}
