package handlers

import "net/http"

func (h *Handlers) GameState(w http.ResponseWriter, r *http.Request) {
	gs, err := h.repo.GetGameState(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, gs)
}

func (h *Handlers) InitGame(w http.ResponseWriter, r *http.Request) {
	if err := h.service.InitGame(r.Context()); err != nil {
		h.error(w, err)
		return
	}
	h.GameState(w, r)
}

func (h *Handlers) NextDay(w http.ResponseWriter, r *http.Request) {
	if err := h.service.NextDay(r.Context()); err != nil {
		h.error(w, err)
		return
	}
	h.GameState(w, r)
}

func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.repo.ListEvents(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, events)
}
