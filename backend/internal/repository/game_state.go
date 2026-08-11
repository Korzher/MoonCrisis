package repository

import (
	"context"

	"MoonCrisis/internal/domain"
)

func (r *Repository) GetGameState(ctx context.Context) (domain.GameState, error) {
	var gs domain.GameState
	err := r.db.QueryRow(ctx,
		`SELECT day, money, rating, game_over FROM game_state WHERE id = 1`,
	).Scan(&gs.Day, &gs.Money, &gs.Rating, &gs.GameOver)
	return gs, err
}

func (r *Repository) UpdateGameState(ctx context.Context, gs domain.GameState) error {
	_, err := r.db.Exec(ctx,
		`UPDATE game_state SET day = $1, money = $2, rating = $3, game_over = $4 WHERE id = 1`,
		gs.Day, gs.Money, gs.Rating, gs.GameOver,
	)
	return err
}

func (r *Repository) UpsertGameState(ctx context.Context, gs domain.GameState) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO game_state (id, day, money, rating, game_over)
		 VALUES (1, $1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE SET
		   day = $1, money = $2, rating = $3, game_over = $4`,
		gs.Day, gs.Money, gs.Rating, gs.GameOver,
	)
	return err
}

// ResetGame — полностью очищает данные прошлой партии и сбрасывает счётчики id.
func (r *Repository) ResetGame(ctx context.Context) error {
	_, err := r.db.Exec(ctx,
		`TRUNCATE deliveries, events, orders, rovers, game_state RESTART IDENTITY CASCADE`)
	return err
}
