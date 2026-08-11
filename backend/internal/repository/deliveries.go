package repository

import (
	"context"

	"MoonCrisis/internal/domain"
)

func (r *Repository) CreateDelivery(ctx context.Context, d domain.Delivery) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO deliveries (rover_id, order_id, started_day) VALUES ($1, $2, $3) RETURNING id`,
		d.RoverID, d.OrderID, d.StartedDay,
	).Scan(&id)
	return id, err
}

// MarkDelivered — груз доставлен (награда начислена), но ровер ещё едет обратно.
func (r *Repository) MarkDelivered(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE deliveries SET result = 'delivered' WHERE id = $1 AND finish_day IS NULL`, id)
	return err
}

func (r *Repository) ListActiveDeliveries(ctx context.Context) ([]domain.Delivery, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, rover_id, order_id, started_day, finish_day, COALESCE(result, ''), duration
		 FROM deliveries WHERE finish_day IS NULL ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []domain.Delivery
	for rows.Next() {
		var d domain.Delivery
		if err := rows.Scan(&d.ID, &d.RoverID, &d.OrderID, &d.StartedDay, &d.FinishDay, &d.Result, &d.Duration); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func (r *Repository) CompleteDelivery(ctx context.Context, id int, finishDay int, result string, duration int) error {
	_, err := r.db.Exec(ctx,
		`UPDATE deliveries SET finish_day = $1, result = $2, duration = $3 WHERE id = $4`,
		finishDay, result, duration, id,
	)
	return err
}
