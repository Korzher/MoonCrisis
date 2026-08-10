package handlers

import (
	"encoding/json"
	"net/http"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
	"MoonCrisis/internal/service"
)

type Handlers struct {
	repo    *repository.Repository
	service *service.Service
}

func New(repo *repository.Repository, svc *service.Service) *Handlers {
	return &Handlers{repo: repo, service: svc}
}

// --- State ---

func (h *Handlers) GameState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gs, err := h.repo.GetGameState(ctx)
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

// --- Rovers ---

func (h *Handlers) ListRovers(w http.ResponseWriter, r *http.Request) {
	rovers, err := h.repo.ListRovers(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, rovers)
}

// --- Orders ---

func (h *Handlers) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.repo.ListOrders(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, orders)
}

// --- Deliveries ---

type assignRequest struct {
	RoverID int `json:"rover_id"`
	OrderID int `json:"order_id"`
}

func (h *Handlers) AssignRover(w http.ResponseWriter, r *http.Request) {
	var req assignRequest
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

// --- Events ---

func (h *Handlers) ListEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.repo.ListEvents(r.Context())
	if err != nil {
		h.error(w, err)
		return
	}
	h.json(w, events)
}

// --- Shop ---

func (h *Handlers) RepairRover(w http.ResponseWriter, r *http.Request) {
	h.shopAction(w, r, func() error {
		rv := &domain.Rover{}
		return json.NewDecoder(r.Body).Decode(rv)
	}, func(id int) error {
		return h.service.RepairRover(r.Context(), id)
	})
}

func (h *Handlers) ChargeRover(w http.ResponseWriter, r *http.Request) {
	h.shopAction(w, r, func() error {
		rv := &domain.Rover{}
		return json.NewDecoder(r.Body).Decode(rv)
	}, func(id int) error {
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

// --- helpers ---

func (h *Handlers) shopAction(w http.ResponseWriter, r *http.Request, dec func() error, act func(int) error) {
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

func (h *Handlers) json(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *Handlers) error(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
