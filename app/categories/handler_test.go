package categories

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mytheresa/go-hiring-challenge/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCategoriesRepository struct {
	mock.Mock
}

func (m *MockCategoriesRepository) GetAllCategories() ([]models.Category, error) {
	args := m.Called()
	return args.Get(0).([]models.Category), args.Error(1)
}

func (m *MockCategoriesRepository) CreateCategory(category *models.Category) error {
	args := m.Called(category)
	return args.Error(0)
}

func TestCategoriesHandler_HandleGet(t *testing.T) {
	mockRepo := new(MockCategoriesRepository)
	handler := NewCategoriesHandler(mockRepo)

	categories := []models.Category{
		{ID: 1, Code: "clothing", Name: "Clothing"},
		{ID: 2, Code: "shoes", Name: "Shoes"},
	}

	t.Run("successful categories list request", func(t *testing.T) {
		mockRepo.On("GetAllCategories").Return(categories, nil)

		req := httptest.NewRequest("GET", "/categories", nil)
		recorder := httptest.NewRecorder()

		handler.HandleGet(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response []models.Category
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Len(t, response, 2)
		assert.Equal(t, "clothing", response[0].Code)
		assert.Equal(t, "Shoes", response[1].Name)
	})
}

func TestCategoriesHandler_HandleCreate(t *testing.T) {
	mockRepo := new(MockCategoriesRepository)
	handler := NewCategoriesHandler(mockRepo)

	t.Run("successful category creation", func(t *testing.T) {
		newCategory := &models.Category{
			Code: "electronics",
			Name: "Electronics",
		}

		mockRepo.On("CreateCategory", newCategory).Return(nil)

		requestBody := map[string]string{
			"code": "electronics",
			"name": "Electronics",
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response models.Category
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "electronics", response.Code)
		assert.Equal(t, "Electronics", response.Name)
	})

	t.Run("category creation with invalid body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/categories", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("category creation with missing fields", func(t *testing.T) {
		requestBody := map[string]string{
			"code": "electronics",
			// name is missing
		}
		body, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/categories", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.HandleCreate(recorder, req)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})
}
