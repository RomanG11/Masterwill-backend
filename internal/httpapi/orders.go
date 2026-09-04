package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"masterwill-backend/internal/models"
	"masterwill-backend/internal/store"
)

type orderItemRequest struct {
	ProductID int64 `json:"productId"`
	Quantity  int   `json:"quantity"`
}

type createOrderRequest struct {
	CustomerName string              `json:"customerName"`
	Phone        string              `json:"phone"`
	Email        string              `json:"email"`
	City         string              `json:"city"`
	Address      string              `json:"address"`
	Comment      string              `json:"comment"`
	Items        []orderItemRequest  `json:"items"`
}

func (a *api) createOrder(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	if strings.TrimSpace(req.CustomerName) == "" || strings.TrimSpace(req.Phone) == "" {
		writeError(w, http.StatusBadRequest, "вкажіть ім'я та телефон")
		return
	}
	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "кошик порожній")
		return
	}

	items := make([]store.NewOrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, store.NewOrderItem{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	order, err := a.store.CreateOrder(r.Context(), store.NewOrder{
		CustomerName: strings.TrimSpace(req.CustomerName),
		Phone:        strings.TrimSpace(req.Phone),
		Email:        strings.TrimSpace(req.Email),
		City:         strings.TrimSpace(req.City),
		Address:      strings.TrimSpace(req.Address),
		Comment:      strings.TrimSpace(req.Comment),
		Items:        items,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (a *api) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	order, err := a.store.GetOrder(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

// startCheckout hands the frontend what it needs to send the shopper to pay.
// With the mock provider this also immediately marks the order paid, since
// there is no real bank round-trip to wait for.
func (a *api) startCheckout(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	order, err := a.store.GetOrder(r.Context(), id)
	if err != nil {
		writeStoreErr(w, err)
		return
	}

	checkout, err := a.payment.StartCheckout(r.Context(), order)
	if err != nil {
		writeError(w, http.StatusBadGateway, "не вдалося ініціювати оплату")
		return
	}

	if a.payment.Name() == "mock" {
		if err := a.store.UpdatePaymentStatus(r.Context(), id, models.PaymentStatusPaid, "mock"); err != nil {
			writeStoreErr(w, err)
			return
		}
	}

	writeJSON(w, http.StatusOK, checkout)
}
