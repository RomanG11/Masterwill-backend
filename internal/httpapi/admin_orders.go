package httpapi

import (
	"net/http"
	"strconv"

	"masterwill-backend/internal/models"
)

func (a *api) adminListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := a.store.ListOrders(r.Context())
	if err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

type updateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status"`
}

var validOrderStatuses = map[models.OrderStatus]bool{
	models.OrderStatusNew: true, models.OrderStatusPaid: true, models.OrderStatusProcessing: true,
	models.OrderStatusShipped: true, models.OrderStatusCompleted: true, models.OrderStatusCancelled: true,
}

func (a *api) adminUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалідний ідентифікатор")
		return
	}
	var req updateOrderStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "невалідний запит")
		return
	}
	if !validOrderStatuses[req.Status] {
		writeError(w, http.StatusBadRequest, "невалідний статус")
		return
	}
	if err := a.store.UpdateOrderStatus(r.Context(), id, req.Status); err != nil {
		writeStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
