package repository

import (
	"context"

	"MoonCrisis/internal/domain"
)

func (r *Repository) ListOrders(ctx context.Context) ([]domain.Order, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, weight, reward, deadline, risk, x, y, status FROM orders ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.Title, &o.Weight, &o.Reward, &o.Deadline, &o.Risk, &o.X, &o.Y, &o.Status); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *Repository) CreateOrder(ctx context.Context, o domain.Order) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO orders (title, weight, reward, deadline, risk, x, y, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		o.Title, o.Weight, o.Reward, o.Deadline, o.Risk, o.X, o.Y, o.Status,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetOrder(ctx context.Context, id int) (domain.Order, error) {
	var o domain.Order
	err := r.db.QueryRow(ctx,
		`SELECT id, title, weight, reward, deadline, risk, x, y, status FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.Title, &o.Weight, &o.Reward, &o.Deadline, &o.Risk, &o.X, &o.Y, &o.Status)
	return o, err
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	return err
}

// ClaimOrder — атомарно занимает заказ (только если был available).
func (r *Repository) ClaimOrder(ctx context.Context, id int) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE orders SET status='active' WHERE id=$1 AND status='available'`, id)
	return tag.RowsAffected() > 0, err
}
