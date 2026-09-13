package httpapi

import (
	"net/http"
	"strconv"

	"masterwill-backend/internal/models"
	"masterwill-backend/internal/store"
)

func (a *api) adminListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := a.store.ListProducts(r.Context(), store.ProductFilter{})
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

type productRequest struct {
	Slug        string `json:"slug"`
	CategoryID  int64  `json:"categoryId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	AgeLabel    string `json:"ageLabel"`
	PhotoURL    string `json:"photoUrl"`
	AccentColor string `json:"accentColor"`
	PriceCents  int64  `json:"priceCents"`
	Currency    string `json:"currency"`
	StockQty    int    `json:"stockQty"`
	IsActive    bool   `json:"isActive"`
}

func (a *api) adminCreateProduct(w http.ResponseWriter, r *http.Request) {
	var req productRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	if req.Currency == "" {
		req.Currency = "UAH"
	}
	created, err := a.store.CreateProduct(r.Context(), models.Product{
		Slug: req.Slug, CategoryID: req.CategoryID, Name: req.Name, Description: req.Description,
		AgeLabel: req.AgeLabel, PhotoURL: req.PhotoURL, AccentColor: req.AccentColor,
		PriceCents: req.PriceCents, Currency: req.Currency, StockQty: req.StockQty, IsActive: req.IsActive,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *api) adminUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	var req productRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	if req.Currency == "" {
		req.Currency = "UAH"
	}
	err = a.store.UpdateProduct(r.Context(), models.Product{
		ID: id, Slug: req.Slug, CategoryID: req.CategoryID, Name: req.Name, Description: req.Description,
		AgeLabel: req.AgeLabel, PhotoURL: req.PhotoURL, AccentColor: req.AccentColor,
		PriceCents: req.PriceCents, Currency: req.Currency, StockQty: req.StockQty, IsActive: req.IsActive,
	})
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (a *api) adminDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	if err := a.store.DeleteProduct(r.Context(), id); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
