package handlers

import "net/http"

func (h *Handlers) ListRovers(w http.ResponseWriter, r *http.Request) {
	rovers, err := h.repo.ListRovers(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, rovers)
}
