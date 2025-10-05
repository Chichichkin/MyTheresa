package category

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
)

type MockCategoryRepo struct {
	ListAllFunc     func(ctx context.Context) ([]models.Category, error)
	CreateFunc      func(ctx context.Context, newCategory models.Category) error
	GetByIDFunc     func(ctx context.Context, id int) (string, error)
	GetByCodeFunc   func(ctx context.Context, code string) (string, error)
	GetProductsFunc func(ctx context.Context, code string) ([]models.Product, error)
}

func (m *MockCategoryRepo) ListAll(ctx context.Context) ([]models.Category, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockCategoryRepo) Create(ctx context.Context, newCategory models.Category) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, newCategory)
	}
	return nil
}

func (m *MockCategoryRepo) GetByID(ctx context.Context, id int) (string, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return "", nil
}

func (m *MockCategoryRepo) GetByCode(ctx context.Context, code string) (string, error) {
	if m.GetByCodeFunc != nil {
		return m.GetByCodeFunc(ctx, code)
	}
	return "", nil
}

func (m *MockCategoryRepo) GetProducts(ctx context.Context, code string) ([]models.Product, error) {
	if m.GetProductsFunc != nil {
		return m.GetProductsFunc(ctx, code)
	}
	return nil, nil
}

func TestHandler_HandleGet(t *testing.T) {
	tests := []struct {
		name           string
		mockCategories []models.Category
		mockError      error
		expectedStatus int
		expectedBody   Response
	}{
		{
			name: "successful get categories",
			mockCategories: []models.Category{
				{ID: 1, Code: "clothing", Name: "Clothing"},
				{ID: 2, Code: "shoes", Name: "Shoes"},
				{ID: 3, Code: "accessories", Name: "Accessories"},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Categories: []Category{
					{Code: "clothing", Name: "Clothing"},
					{Code: "shoes", Name: "Shoes"},
					{Code: "accessories", Name: "Accessories"},
				},
			},
		},
		{
			name:           "empty categories list",
			mockCategories: []models.Category{},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody: Response{
				Categories: []Category{},
			},
		},
		{
			name:           "database error",
			mockCategories: nil,
			mockError:      errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   Response{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockCategoryRepo{
				ListAllFunc: func(ctx context.Context) ([]models.Category, error) {
					return tt.mockCategories, tt.mockError
				},
			}

			handler := NewCategoryHandler(mockRepo)

			req := httptest.NewRequest("GET", "/categories", nil)
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

func TestHandler_HandlePost(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    CreateRequest
		mockError      error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful create category",
			requestBody: CreateRequest{
				Code: "electronics",
				Name: "Electronics",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "duplicate category code",
			requestBody: CreateRequest{
				Code: "clothing",
				Name: "Clothing",
			},
			mockError:      errors.New("category code already exists"),
			expectedStatus: http.StatusConflict,
			expectedError:  "category code already exists",
		},
		{
			name: "database error",
			requestBody: CreateRequest{
				Code: "electronics",
				Name: "Electronics",
			},
			mockError:      errors.New("database connection failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "database connection failed",
		},
		{
			name: "missing code",
			requestBody: CreateRequest{
				Name: "Electronics",
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Code and name are required",
		},
		{
			name: "missing name",
			requestBody: CreateRequest{
				Code: "electronics",
			},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Code and name are required",
		},
		{
			name:           "invalid JSON",
			requestBody:    CreateRequest{},
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockCategoryRepo{}
			if tt.name != "invalid JSON" && tt.name != "missing code" && tt.name != "missing name" {
				mockRepo.CreateFunc = func(ctx context.Context, newCategory models.Category) error {
					return tt.mockError
				}
			}

			handler := NewCategoryHandler(mockRepo)

			var req *http.Request
			if tt.name == "invalid JSON" {
				req = httptest.NewRequest("POST", "/categories", bytes.NewBufferString("invalid json"))
			} else {
				body, _ := json.Marshal(tt.requestBody)
				req = httptest.NewRequest("POST", "/categories", bytes.NewBuffer(body))
			}
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.HandlePost(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var errorResponse map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse["error"])
			} else if tt.expectedStatus == http.StatusOK {
				var successResponse map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &successResponse)
				assert.NoError(t, err)
				assert.Equal(t, "Category created successfully", successResponse["message"])
			}
		})
	}
}
