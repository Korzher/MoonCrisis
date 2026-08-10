package handlers

import (
	"encoding/json"
	"net/http"

	"MoonCrisis/internal/domain"
)

func (h *Handlers) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.repo.ListOrders(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, orders)
}

func (h *Handlers) AssignRover(w http.ResponseWriter, r *http.Request) {
	var req domain.AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, err)
		return
	}
	if err := h.service.AssignRover(r.Context(), req.RoverID, req.OrderID); err != nil {
		h.error(w, err)
		return
	}
	h.json(w, map[string]string{"status": "assigned"})
}
