package httpapi

import (
	"net/http"

	"masterwill-backend/internal/auth"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func (a *api) adminLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}

	admin, err := a.store.GetAdminByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "невірний email або пароль")
		return
	}
	if !auth.CheckPassword(admin.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "невірний email або пароль")
		return
	}

	token, err := auth.IssueToken(a.cfg.JWTSecret, admin.ID, admin.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "не вдалося створити токен")
		return
	}
	writeJSON(w, http.StatusOK, loginResponse{Token: token, Email: admin.Email})
}
