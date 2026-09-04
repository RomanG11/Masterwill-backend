package httpapi

import (
	"net/http"

	"masterwill-backend/internal/models"
	"masterwill-backend/internal/payment"
)

// liqpayCallback handles LiqPay's server-to-server "server_url" POST. It's a
// no-op (200 OK, ignored) unless the active provider is actually LiqPay —
// safe to leave mounted regardless of PAYMENT_PROVIDER.
func (a *api) liqpayCallback(w http.ResponseWriter, r *http.Request) {
	lp, ok := a.payment.(*payment.LiqPay)
	if !ok {
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	orderID, status, err := lp.VerifyCallback(r.FormValue("data"), r.FormValue("signature"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний підпис")
		return
	}

	paymentStatus := models.PaymentStatusFailed
	if status == "success" || status == "sandbox" {
		paymentStatus = models.PaymentStatusPaid
	}
	if err := a.store.UpdatePaymentStatus(r.Context(), orderID, paymentStatus, "liqpay"); err != nil {
		writeStoreErr(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
