package httpapi

import (
	"net/http"
	"strconv"

	"masterwill-backend/internal/models"
)

func (a *api) adminListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := a.store.ListCategories(r.Context())
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

type categoryRequest struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

func (a *api) adminCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	created, err := a.store.CreateCategory(r.Context(), models.Category{Slug: req.Slug, Name: req.Name, SortOrder: req.SortOrder})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (a *api) adminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	var req categoryRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	if err := a.store.UpdateCategory(r.Context(), models.Category{ID: id, Slug: req.Slug, Name: req.Name, SortOrder: req.SortOrder}); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}

func (a *api) adminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	if err := a.store.DeleteCategory(r.Context(), id); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
