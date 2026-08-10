package di

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"MoonCrisis/internal/datasource"
	"MoonCrisis/internal/handlers"
	"MoonCrisis/internal/repository"
	"MoonCrisis/internal/service"
)

// Build собирает DI-граф приложения.
func Build() *fx.App {
	return fx.New(
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
}

// NewRouter — регистрирует все маршруты
func NewRouter(h *handlers.Handlers) http.Handler {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// NewServer — http-сервер (порт берём из конфига)
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
