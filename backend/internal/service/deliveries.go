package service

import (
	"context"
	"math/rand"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// advanceDeliveries — обрабатывает все активные доставки за день.
func advanceDeliveries(ctx context.Context, t *repository.Repository, state *domain.GameState) error {
	deliveries, err := t.ListActiveDeliveries(ctx)
	if err != nil {
		return err
	}
	for _, d := range deliveries {
		if err := advanceOneDelivery(ctx, t, state, d); err != nil {
			return err
		}
	}
	return nil
}

// advanceOneDelivery — движение, доезд, разворот или возврат для одной доставки.
func advanceOneDelivery(ctx context.Context, t *repository.Repository, state *domain.GameState, d domain.Delivery) error {
	rv, err := t.GetRover(ctx, d.RoverID)
	if err != nil {
		return err
	}
	o, err := t.GetOrder(ctx, d.OrderID)
	if err != nil {
		return err
	}

	if !isDelivered(d) {
		// Едем к цели (с грузом медленнее)
		if err := moveRoverTo(ctx, t, &rv, o.X, o.Y, roverSpeed(rv, o.Weight), batteryPerCell(o.Weight)); err != nil {
			return err
		}
		if rv.X == o.X && rv.Y == o.Y {
			return settleDelivery(ctx, t, state, &rv, o, d)
		}
		if o.Deadline < state.Day {
			return turnBack(ctx, t, state, o, d)
		}
		return nil
	}

	// Возврат к базе (пустой — полная скорость)
	if err := moveRoverTo(ctx, t, &rv, BaseX, BaseY, rv.Speed, batteryPerCell(0)); err != nil {
		return err
	}
	if rv.X == BaseX && rv.Y == BaseY {
		return finishReturn(ctx, t, state, &rv, o, d)
	}
	return nil
}

// roverSpeed — скорость с грузом: −1 за каждые 40 кг (минимум 1).
func roverSpeed(rv domain.Rover, weight int) int {
	speed := rv.Speed - weight/40
	if speed < 1 {
		return 1
	}
	return speed
}

// moveRoverTo — сдвигает ровер на speed клеток и списывает батарею за реальный путь.
func moveRoverTo(ctx context.Context, t *repository.Repository, rv *domain.Rover, tx, ty, speed int, perCell float64) error {
	nx, ny := moveToward(rv.X, rv.Y, tx, ty, speed)
	if nx == rv.X && ny == rv.Y {
		return nil
	}
	dx, dy := nx-rv.X, ny-rv.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	cells := dx + dy
	rv.Battery -= int(perCell * float64(cells))
	if rv.Battery < 0 {
		rv.Battery = 0
	}
	rv.X, rv.Y = nx, ny
	return t.UpdateRover(ctx, *rv)
}

// settleDelivery — доехали до точки: провал, опоздание или успех.
func settleDelivery(ctx context.Context, t *repository.Repository, state *domain.GameState, rv *domain.Rover, o domain.Order, d domain.Delivery) error {
	if rand.Intn(100) < o.Risk {
		// Провал: ровер сломался
		state.Rating -= 10
		rv.Status = "broken"
		rv.X, rv.Y = BaseX, BaseY
		_ = t.UpdateRover(ctx, *rv)
		_ = t.UpdateOrderStatus(ctx, o.ID, "failed")
		_ = t.CompleteDelivery(ctx, d.ID, state.Day, "failed", 0)
		return t.AddEvent(ctx, state.Day, "Доставка «"+o.Title+"» провалена: ровер сломан, рейтинг -10")
	}
	if o.Deadline < state.Day {
		// Опоздали: приехали, но срок прошёл
		state.Rating -= 5
		_ = t.UpdateOrderStatus(ctx, o.ID, "expired")
		_ = t.MarkDelivered(ctx, d.ID)
		return t.AddEvent(ctx, state.Day, "Заказ «"+o.Title+"» доставлен слишком поздно: без награды, рейтинг -5")
	}
	// Успех
	state.Money += o.Reward
	state.Rating += 5
	_ = t.UpdateOrderStatus(ctx, o.ID, "completed")
	_ = t.MarkDelivered(ctx, d.ID)
	return t.AddEvent(ctx, state.Day, "Доставка «"+o.Title+"» выполнена: +"+itoa(o.Reward)+"₽, рейтинг +5")
}

// turnBack — разворот: не доехали, а дедлайн прошёл.
func turnBack(ctx context.Context, t *repository.Repository, state *domain.GameState, o domain.Order, d domain.Delivery) error {
	state.Rating -= 5
	_ = t.UpdateOrderStatus(ctx, o.ID, "expired")
	_ = t.MarkDelivered(ctx, d.ID)
	return t.AddEvent(ctx, state.Day, "Заказ «"+o.Title+"» просрочен — ровер развернулся и возвращается, рейтинг -5")
}

// finishReturn — вернулись на базу: ровер свободен, доставка закрыта.
func finishReturn(ctx context.Context, t *repository.Repository, state *domain.GameState, rv *domain.Rover, o domain.Order, d domain.Delivery) error {
	result := "success"
	if o.Status != "completed" {
		result = "failed"
	}
	rv.Status = "idle"
	_ = t.UpdateRover(ctx, *rv)
	_ = t.CompleteDelivery(ctx, d.ID, state.Day, result, 0)
	return t.AddEvent(ctx, state.Day, "Ровер вернулся на базу")
}
