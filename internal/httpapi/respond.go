package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"masterwill-backend/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// writeStoreErr maps a store-layer error to the right HTTP status: a missing
// row is a 404, anything else is treated as unexpected and logged.
func writeStoreErr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "не знайдено")
		return
	}
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "внутрішня помилка сервера")
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
