// Package httpapi wires the storefront + admin REST API: thin handlers that
// decode requests, call into store, and encode responses. No business logic
// lives here beyond request validation — that stays in store/payment.
package httpapi

import (
	"net/http"

	"masterwill-backend/internal/config"
	"masterwill-backend/internal/payment"
	"masterwill-backend/internal/store"
)

type api struct {
	store   *store.Store
	cfg     config.Config
	payment payment.Provider
}

func NewRouter(s *store.Store, cfg config.Config, p payment.Provider) http.Handler {
	a := &api{store: s, cfg: cfg, payment: p}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", a.health)

	// Product photos (admin-uploaded and seed alike) — public read, no auth,
	// same as any other static asset a storefront needs to display.
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadsDir))))
	mux.HandleFunc("POST /api/admin/uploads", a.requireAdmin(a.adminUploadPhoto))

	// Public storefront
	mux.HandleFunc("GET /api/categories", a.listCategories)
	mux.HandleFunc("GET /api/products", a.listProducts)
	mux.HandleFunc("GET /api/products/{slug}", a.getProduct)
	mux.HandleFunc("POST /api/orders", a.createOrder)
	mux.HandleFunc("GET /api/orders/{id}", a.getOrder)
	mux.HandleFunc("POST /api/orders/{id}/checkout", a.startCheckout)
	mux.HandleFunc("POST /api/payments/liqpay/callback", a.liqpayCallback)

	// Admin auth
	mux.HandleFunc("POST /api/admin/login", a.adminLogin)

	// Admin (protected)
	mux.HandleFunc("GET /api/admin/products", a.requireAdmin(a.adminListProducts))
	mux.HandleFunc("POST /api/admin/products", a.requireAdmin(a.adminCreateProduct))
	mux.HandleFunc("PUT /api/admin/products/{id}", a.requireAdmin(a.adminUpdateProduct))
	mux.HandleFunc("DELETE /api/admin/products/{id}", a.requireAdmin(a.adminDeleteProduct))

	mux.HandleFunc("GET /api/admin/categories", a.requireAdmin(a.adminListCategories))
	mux.HandleFunc("POST /api/admin/categories", a.requireAdmin(a.adminCreateCategory))
	mux.HandleFunc("PUT /api/admin/categories/{id}", a.requireAdmin(a.adminUpdateCategory))
	mux.HandleFunc("DELETE /api/admin/categories/{id}", a.requireAdmin(a.adminDeleteCategory))

	mux.HandleFunc("GET /api/admin/orders", a.requireAdmin(a.adminListOrders))
	mux.HandleFunc("PATCH /api/admin/orders/{id}/status", a.requireAdmin(a.adminUpdateOrderStatus))

	var handler http.Handler = mux
	handler = withCORS(cfg.CORSOrigin)(handler)
	handler = withLogging(handler)
	handler = withRecover(handler)
	return handler
}

func (a *api) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
