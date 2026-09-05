package httpapi

import (
	"net/http"
	"strconv"
)

type productMediaRequest struct {
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
}

func (a *api) adminListProductMedia(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	media, err := a.store.ListProductMedia(r.Context(), productID)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, media)
}

func (a *api) adminAddProductMedia(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	var req productMediaRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	if req.MediaType != "photo" && req.MediaType != "video" {
		writeError(w, http.StatusBadRequest, "тип медіа має бути photo або video")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "потрібен url")
		return
	}
	media, err := a.store.AddProductMedia(r.Context(), productID, req.MediaType, req.URL)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, media)
}

func (a *api) adminDeleteProductMedia(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор товару")
		return
	}
	mediaID, err := strconv.ParseInt(r.PathValue("mediaId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор медіа")
		return
	}
	if err := a.store.DeleteProductMedia(r.Context(), productID, mediaID); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
