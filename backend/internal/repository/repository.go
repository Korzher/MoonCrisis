package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"MoonCrisis/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// --- GameState ---

func (r *Repository) GetGameState(ctx context.Context) (domain.GameState, error) {
	var gs domain.GameState
	err := r.pool.QueryRow(ctx,
		`SELECT day, money, rating, game_over FROM game_state WHERE id = 1`,
	).Scan(&gs.Day, &gs.Money, &gs.Rating, &gs.GameOver)
	return gs, err
}

func (r *Repository) UpdateGameState(ctx context.Context, gs domain.GameState) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE game_state SET day = $1, money = $2, rating = $3, game_over = $4 WHERE id = 1`,
		gs.Day, gs.Money, gs.Rating, gs.GameOver,
	)
	return err
}

// --- Rovers ---

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

// --- Orders ---

func (r *Repository) ListOrders(ctx context.Context) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx,
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
	err := r.pool.QueryRow(ctx,
		`INSERT INTO orders (title, weight, reward, deadline, risk, x, y, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		o.Title, o.Weight, o.Reward, o.Deadline, o.Risk, o.X, o.Y, o.Status,
	).Scan(&id)
	return id, err
}

func (r *Repository) GetOrder(ctx context.Context, id int) (domain.Order, error) {
	var o domain.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, weight, reward, deadline, risk, x, y, status FROM orders WHERE id = $1`, id,
	).Scan(&o.ID, &o.Title, &o.Weight, &o.Reward, &o.Deadline, &o.Risk, &o.X, &o.Y, &o.Status)
	return o, err
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, id int, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	return err
}

// --- Deliveries ---

func (r *Repository) CreateDelivery(ctx context.Context, d domain.Delivery) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO deliveries (rover_id, order_id, started_day) VALUES ($1, $2, $3) RETURNING id`,
		d.RoverID, d.OrderID, d.StartedDay,
	).Scan(&id)
	return id, err
}

func (r *Repository) ListActiveDeliveries(ctx context.Context) ([]domain.Delivery, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, rover_id, order_id, started_day, finish_day, result, duration
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
	_, err := r.pool.Exec(ctx,
		`UPDATE deliveries SET finish_day = $1, result = $2, duration = $3 WHERE id = $4`,
		finishDay, result, duration, id,
	)
	return err
}

// --- Events ---

func (r *Repository) AddEvent(ctx context.Context, day int, message string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO events (day, message) VALUES ($1, $2)`, day, message)
	return err
}

func (r *Repository) ListEvents(ctx context.Context) ([]domain.Event, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, day, message FROM events ORDER BY id DESC LIMIT 50`)
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
