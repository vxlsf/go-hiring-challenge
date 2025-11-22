package catalog

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProductsRepository struct {
	mock.Mock
}

func (m *MockProductsRepository) GetProductsWithFilters(categoryCode string, priceLessThan *float64, offset, limit int) ([]models.Product, int64, error) {
	args := m.Called(categoryCode, priceLessThan, offset, limit)
	return args.Get(0).([]models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductsRepository) GetProductByCode(code string) (*models.Product, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func TestCatalogHandler_HandleGet(t *testing.T) {
	mockRepo := new(MockProductsRepository)
	handler := NewCatalogHandler(mockRepo)

	products := []models.Product{
		{
			Code:  "PROD001",
			Price: NewDecimal(10.99),
			Category: models.Category{
				Code: "clothing",
				Name: "Clothing",
			},
		},
	}

	t.Run("successful request with default pagination", func(t *testing.T) {
		mockRepo.On("GetProductsWithFilters", "", (*float64)(nil), 0, 10).
			Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response CatalogResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), response.Total)
		assert.Equal(t, 0, response.Offset)
		assert.Equal(t, 10, response.Limit)
		assert.Len(t, response.Products, 1)
		assert.Equal(t, "PROD001", response.Products[0].Code)
		assert.Equal(t, "clothing", response.Products[0].Category.Code)
	})

	t.Run("successful request with custom pagination", func(t *testing.T) {
		mockRepo.On("GetProductsWithFilters", "", (*float64)(nil), 10, 20).
			Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog?offset=10&limit=20", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response CatalogResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, 10, response.Offset)
		assert.Equal(t, 20, response.Limit)
	})

	t.Run("Invalid offset parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/catalog?offset=-1&limit=10", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid offset parameter, must be a non-negative integer", errorResponse["error"])

		mockRepo.AssertNotCalled(t, "GetProductsWithFilters")
	})

	t.Run("Invalid limit parameter", func(t *testing.T) {
		mockRepo.On("GetProductsWithFilters", "", (*float64)(nil), 0, 101).Return(nil, int64(0), assert.AnError)

		req := httptest.NewRequest("GET", "/catalog?offset=0&limit=101", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var errorResponse map[string]string
		err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid limit parameter, must be an integer between 1 and 100", errorResponse["error"])

		mockRepo.AssertNotCalled(t, "GetProductsWithFilters")
	})

	t.Run("successful request with category filter", func(t *testing.T) {
		category := "clothing"
		mockRepo.On("GetProductsWithFilters", category, (*float64)(nil), 0, 10).
			Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog?category=clothing", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockRepo.AssertCalled(t, "GetProductsWithFilters", category, (*float64)(nil), 0, 10)
	})

	t.Run("successful request with price filter", func(t *testing.T) {
		price := 50.0
		mockRepo.On("GetProductsWithFilters", "", &price, 0, 10).
			Return(products, int64(1), nil)

		req := httptest.NewRequest("GET", "/catalog?price_less_than=50", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockRepo.AssertCalled(t, "GetProductsWithFilters", "", &price, 0, 10)
	})
}

func TestCatalogHandler_HandleGetDetail(t *testing.T) {
	mockRepo := new(MockProductsRepository)
	handler := NewCatalogHandler(mockRepo)

	product := &models.Product{
		Code:  "PROD001",
		Price: NewDecimal(10.99),
		Category: models.Category{
			Code: "clothing",
			Name: "Clothing",
		},
		Variants: []models.Variant{
			{
				Name:  "Variant A",
				SKU:   "SKU001A",
				Price: NewDecimal(11.99),
			},
			{
				Name:  "Variant B",
				SKU:   "SKU001B",
				Price: NewDecimal(0),
			},
		},
	}

	t.Run("successful product detail request", func(t *testing.T) {
		mockRepo.On("GetProductByCode", "PROD001").Return(product, nil)

		req := httptest.NewRequest("GET", "/catalog/PROD001?code=PROD001", nil)
		recorder := httptest.NewRecorder()
		handler.HandleGetDetail(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)

		var response ProductDetailResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "PROD001", response.Code)
		assert.Equal(t, "clothing", response.Category.Code)
		assert.Len(t, response.Variants, 2)
		assert.Equal(t, 11.99, response.Variants[0].Price)
		assert.Equal(t, 10.99, response.Variants[1].Price)
	})

	t.Run("product not found", func(t *testing.T) {
		mockRepo.On("GetProductByCode", "NONEXISTENT").Return(nil, assert.AnError)

		req := httptest.NewRequest("GET", "/catalog/NONEXISTENT?code=NONEXISTENT", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGetDetail(recorder, req)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
}

func NewDecimal(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}
