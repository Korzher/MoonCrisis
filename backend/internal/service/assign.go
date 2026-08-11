package service

import (
	"context"
	"errors"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// AssignRover — назначить ровер на заказ.
func (s *Service) AssignRover(ctx context.Context, roverID, orderID int) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		rv, err := t.GetRover(ctx, roverID)
		if err != nil {
			return err
		}
		o, err := t.GetOrder(ctx, orderID)
		if err != nil {
			return err
		}

		// Валидации
		if rv.Status != "idle" {
			return errors.New("ровер занят")
		}
		if o.Status != "available" {
			return errors.New("заказ недоступен")
		}
		if o.Weight > rv.Capacity {
			return errors.New("превышена грузоподъёмность")
		}
		if batteryCost(o.X, o.Y, o.Weight) > float64(rv.Battery) {
			return errors.New("не хватает батареи на маршрут")
		}

		// Атомарно занимаем ровер и заказ (защита от гонки)
		ok, err := t.ClaimRover(ctx, rv.ID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("ровер занят")
		}
		ok, err = t.ClaimOrder(ctx, o.ID)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("заказ недоступен")
		}

		// Создаём доставку (в той же транзакции)
		gs, err := t.GetGameStateForUpdate(ctx)
		if err != nil {
			return err
		}
		_, err = t.CreateDelivery(ctx, domain.Delivery{
			RoverID:    rv.ID,
			OrderID:    o.ID,
			StartedDay: gs.Day,
		})
		return err
	})
}
