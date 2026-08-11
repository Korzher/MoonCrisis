package handlers

import (
	"encoding/json"
	"net/http"
)

func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/game/state", h.GameState)
	mux.HandleFunc("POST /api/game/start", h.InitGame)
	mux.HandleFunc("POST /api/game/next-day", h.NextDay)

	mux.HandleFunc("GET /api/rovers", h.ListRovers)
	mux.HandleFunc("GET /api/orders", h.ListOrders)
	mux.HandleFunc("POST /api/deliveries", h.AssignRover)
	mux.HandleFunc("GET /api/events", h.ListEvents)

	mux.HandleFunc("POST /api/shop/repair", h.RepairRover)
	mux.HandleFunc("POST /api/shop/charge", h.ChargeRover)
	mux.HandleFunc("POST /api/shop/buy", h.BuyRover)

	mux.HandleFunc("GET /api/deliveries", h.ListDeliveries)
}
