package service

import (
	"context"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// handleOrders — просрочка заказов + пополнение available до 5.
func handleOrders(ctx context.Context, t *repository.Repository, state *domain.GameState, day int) error {
	orders, err := t.ListOrders(ctx)
	if err != nil {
		return err
	}

	avail := 0
	for _, o := range orders {
		if o.Status == "active" && o.Deadline < day {
			state.Rating -= 5
			_ = t.UpdateOrderStatus(ctx, o.ID, "expired")
			_ = t.AddEvent(ctx, day, "Заказ «"+o.Title+"» просрочен: рейтинг -5")
		} else if o.Status == "available" {
			avail++
			if o.Deadline < day {
				state.Rating -= 3
				_ = t.UpdateOrderStatus(ctx, o.ID, "expired")
				_ = t.AddEvent(ctx, day, "Заказ «"+o.Title+"» не взят вовремя: рейтинг -3")
			}
		}
	}

	for avail < 5 {
		generateOrder(ctx, t, day)
		avail++
	}
	return nil
}
