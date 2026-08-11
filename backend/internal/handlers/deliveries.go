package handlers

import "net/http"

func (h *Handlers) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	deliveries, err := h.repo.ListActiveDeliveries(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, deliveries)
}
