package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"MoonCrisis/internal/datasource"
	"MoonCrisis/internal/handlers"
	"MoonCrisis/internal/repository"
	"MoonCrisis/internal/service"
)

func main() {
	app := fx.New(
		fx.Provide(
			datasource.Load,
			func() (context.Context, func(), error) {
				return context.Background(), func() {}, nil
			},
			func(ctx context.Context, cfg datasource.Config) (*pgxpool.Pool, error) {
				return datasource.New(ctx, cfg.DatabaseURL)
			},
			repository.New,
			service.New,
			handlers.New,
			NewRouter,
			NewServer,
		),
		fx.Invoke(func(*http.Server) {}),
	)

	app.Run()
}

// NewRouter — регистрирует все маршруты
func NewRouter(h *handlers.Handlers) http.Handler {
	mux := http.NewServeMux()

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

	return mux
}

func NewServer(lc fx.Lifecycle, cfg datasource.Config, h http.Handler) *http.Server {
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: h}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go srv.ListenAndServe()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}
