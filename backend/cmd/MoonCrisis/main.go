package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"MoonCrisis/internal/datasource"
)

func main() {
	app := fx.New(
		fx.Provide(
			datasource.Load,
			func(cfg datasource.Config) (context.Context, func(), error) {
				return context.Background(), func() {}, nil
			},
			func(ctx context.Context, cfg datasource.Config) (*pgxpool.Pool, error) {
				return datasource.New(ctx, cfg.DatabaseURL)
			},
			func() http.Handler {
				mux := http.NewServeMux()
				mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
				})
				return mux
			},
			func(lc fx.Lifecycle, h http.Handler) *http.Server {
				srv := &http.Server{Addr: ":8080", Handler: h}
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
			},
		),
		fx.Invoke(func(*http.Server) {}),
	)

	app.Run()
}
