package repository

import (
	"context"

	"MoonCrisis/internal/domain"
)

func (r *Repository) AddEvent(ctx context.Context, day int, message string) error {
	_, err := r.q.Exec(ctx, `INSERT INTO events (day, message) VALUES ($1, $2)`, day, message)
	return err
}

func (r *Repository) ListEvents(ctx context.Context) ([]domain.Event, error) {
	rows, err := r.q.Query(ctx, `SELECT id, day, message FROM events ORDER BY id DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.Event
	for rows.Next() {
		var e domain.Event
		if err := rows.Scan(&e.ID, &e.Day, &e.Message); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
