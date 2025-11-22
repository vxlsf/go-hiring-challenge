package catalog

import (
	"net/http"
	"strconv"

	"github.com/mytheresa/go-hiring-challenge/app/api"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CatalogHandler struct {
	repo models.ProductsRepositoryInterface
}

func NewCatalogHandler(r models.ProductsRepositoryInterface) *CatalogHandler {
	return &CatalogHandler{
		repo: r,
	}
}

type CatalogResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int64             `json:"total"`
	Offset   int               `json:"offset"`
	Limit    int               `json:"limit"`
}

type ProductResponse struct {
	Code     string           `json:"code"`
	Price    float64          `json:"price"`
	Category CategoryResponse `json:"category"`
}

type CategoryResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type ProductDetailResponse struct {
	Code     string                 `json:"code"`
	Price    float64                `json:"price"`
	Category CategoryResponse       `json:"category"`
	Variants []ProductVariantDetail `json:"variants"`
}

type ProductVariantDetail struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
}

func (h *CatalogHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	offsetStr := query.Get("offset")
	limitStr := query.Get("limit")
	offset := 0
	limit := 10
	var offsetErr error
	var limitErr error

	if offsetStr != "" {
		offset, offsetErr = strconv.Atoi(offsetStr)
		if offsetErr != nil || offset < 0 {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter, must be a non-negative integer")
			return
		}
	}

	if limitStr != "" {
		limit, limitErr = strconv.Atoi(limitStr)
		if limitErr != nil || limit < 1 || limit > 100 {
			api.ErrorResponse(w, http.StatusBadRequest, "Invalid limit parameter, must be an integer between 1 and 100")
			return
		}
	}

	categoryCode := query.Get("category")
	var priceLessThan *float64
	if priceStr := query.Get("price_less_than"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			priceLessThan = &price
		}
	}

	products, total, err := h.repo.GetProductsWithFilters(categoryCode, priceLessThan, offset, limit)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "Failed to fetch products")
		return
	}

	productResponses := make([]ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = ProductResponse{
			Code:  product.Code,
			Price: product.Price.InexactFloat64(),
			Category: CategoryResponse{
				Code: product.Category.Code,
				Name: product.Category.Name,
			},
		}
	}

	response := CatalogResponse{
		Products: productResponses,
		Total:    total,
		Offset:   offset,
		Limit:    limit,
	}

	api.OKResponse(w, response)
}

func (h *CatalogHandler) HandleGetDetail(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		code = r.URL.Query().Get("code")
	}

	if code == "" {
		api.ErrorResponse(w, http.StatusBadRequest, "Product code is required")
		return
	}

	product, err := h.repo.GetProductByCode(code)
	if err != nil {
		api.ErrorResponse(w, http.StatusNotFound, "Product not found")
		return
	}

	variants := make([]ProductVariantDetail, len(product.Variants))
	for i, variant := range product.Variants {
		price := variant.Price.InexactFloat64()
		if variant.Price.IsZero() {
			price = product.Price.InexactFloat64()
		}

		variants[i] = ProductVariantDetail{
			Name:  variant.Name,
			SKU:   variant.SKU,
			Price: price,
		}
	}

	response := ProductDetailResponse{
		Code:  product.Code,
		Price: product.Price.InexactFloat64(),
		Category: CategoryResponse{
			Code: product.Category.Code,
			Name: product.Category.Name,
		},
		Variants: variants,
	}

	api.OKResponse(w, response)
}
