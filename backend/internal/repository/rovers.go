package repository

import (
	"context"

	"MoonCrisis/internal/domain"
)

func (r *Repository) ListRovers(ctx context.Context) ([]domain.Rover, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, battery, capacity, speed, status, x, y FROM rovers ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rovers []domain.Rover
	for rows.Next() {
		var rv domain.Rover
		if err := rows.Scan(&rv.ID, &rv.Name, &rv.Battery, &rv.Capacity, &rv.Speed, &rv.Status, &rv.X, &rv.Y); err != nil {
			return nil, err
		}
		rovers = append(rovers, rv)
	}
	return rovers, rows.Err()
}

func (r *Repository) CreateRover(ctx context.Context, rv domain.Rover) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO rovers (name, battery, capacity, speed, status, x, y)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		rv.Name, rv.Battery, rv.Capacity, rv.Speed, rv.Status, rv.X, rv.Y,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetRover(ctx context.Context, id int) (domain.Rover, error) {
	var rv domain.Rover
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, battery, capacity, speed, status, x, y FROM rovers WHERE id = $1`, id,
	).Scan(&rv.ID, &rv.Name, &rv.Battery, &rv.Capacity, &rv.Speed, &rv.Status, &rv.X, &rv.Y)
	return rv, err
}

func (r *Repository) UpdateRover(ctx context.Context, rv domain.Rover) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE rovers SET battery = $1, status = $2, x = $3, y = $4 WHERE id = $5`,
		rv.Battery, rv.Status, rv.X, rv.Y, rv.ID,
	)
	return err
}

// ClaimRover — атомарно занимает ровер (только если был idle).
func (r *Repository) ClaimRover(ctx context.Context, id int) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE rovers SET status='on_mission' WHERE id=$1 AND status='idle'`, id)
	return tag.RowsAffected() > 0, err
}
