package handlers

import (
	"encoding/json"
	"net/http"

	"MoonCrisis/internal/domain"
)

func (h *Handlers) RepairRover(w http.ResponseWriter, r *http.Request) {
	h.shopAction(w, r, func(id int) error {
		return h.service.RepairRover(r.Context(), id)
	})
}

func (h *Handlers) ChargeRover(w http.ResponseWriter, r *http.Request) {
	h.shopAction(w, r, func(id int) error {
		return h.service.ChargeRover(r.Context(), id)
	})
}

func (h *Handlers) BuyRover(w http.ResponseWriter, r *http.Request) {
	if err := h.service.BuyRover(r.Context()); err != nil {
		h.error(w, err)
		return
	}
	h.json(w, map[string]string{"status": "bought"})
}

func (h *Handlers) shopAction(w http.ResponseWriter, r *http.Request, act func(int) error) {
	rv := &domain.Rover{}
	if err := json.NewDecoder(r.Body).Decode(rv); err != nil {
		h.error(w, err)
		return
	}
	if err := act(rv.ID); err != nil {
		h.error(w, err)
		return
	}
	h.json(w, map[string]string{"status": "ok"})
}
