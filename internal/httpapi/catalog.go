package httpapi

import (
	"net/http"

	"masterwill-backend/internal/store"
)

func (a *api) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := a.store.ListCategories(r.Context())
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (a *api) listProducts(w http.ResponseWriter, r *http.Request) {
	f := store.ProductFilter{
		CategorySlug: r.URL.Query().Get("category"),
		Search:       r.URL.Query().Get("q"),
		ActiveOnly:   true,
	}
	products, err := a.store.ListProducts(r.Context(), f)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, products)
}

func (a *api) getProduct(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	p, err := a.store.GetProductBySlug(r.Context(), slug)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}
