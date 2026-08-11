package service

import (
	"context"
	"errors"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// NextDay — завершить день: доставки, заказы, конец игры.
func (s *Service) NextDay(ctx context.Context) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		state, err := t.GetGameStateForUpdate(ctx)
		if err != nil {
			return err
		}
		if state.GameOver {
			return errors.New("игра окончена")
		}

		state.Day++

		if err := advanceDeliveries(ctx, t, &state); err != nil {
			return err
		}
		if err := handleOrders(ctx, t, &state, state.Day); err != nil {
			return err
		}
		if err := finalizeDay(ctx, t, &state, state.Day); err != nil {
			return err
		}

		return t.UpdateGameState(ctx, state)
	})
}

// finalizeDay — проверяет конец игры (поражение по рейтингу или победа на 50-й день).
func finalizeDay(ctx context.Context, t *repository.Repository, state *domain.GameState, day int) error {
	if day >= 50 {
		state.GameOver = true
		if err := t.AddEvent(ctx, day, "Игра завершена: вы победили, заработав "+itoa(state.Money)+" монет"); err != nil {
			return err
		}
	} else if state.Rating <= 0 {
		state.GameOver = true
		if err := t.AddEvent(ctx, day, "Рейтинг базы упал до нуля — игра окончена"); err != nil {
			return err
		}
	}
	return nil
}

func isDelivered(d domain.Delivery) bool { return d.Result == "delivered" }
