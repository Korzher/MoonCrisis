package service

import (
	"context"
	"errors"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// RepairRover — починить сломанный ровер.
func (s *Service) RepairRover(ctx context.Context, roverID int) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameStateForUpdate(ctx)
		if err != nil {
			return err
		}
		if gs.Money < 100 {
			return errors.New("недостаточно средств")
		}
		rv, err := t.GetRover(ctx, roverID)
		if err != nil {
			return err
		}
		if rv.Status != "broken" {
			return errors.New("ровер не сломан")
		}
		gs.Money -= 100
		rv.Status = "idle"
		if err := t.UpdateRover(ctx, rv); err != nil {
			return err
		}
		if err := t.UpdateGameState(ctx, gs); err != nil {
			return err
		}
		return t.AddEvent(ctx, gs.Day, "Ровер отремонтирован: -100₽")
	})
}

// ChargeRover — полная зарядка батареи.
func (s *Service) ChargeRover(ctx context.Context, roverID int) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameStateForUpdate(ctx)
		if err != nil {
			return err
		}
		if gs.Money < 50 {
			return errors.New("недостаточно средств")
		}
		rv, err := t.GetRover(ctx, roverID)
		if err != nil {
			return err
		}
		if rv.Status != "idle" {
			return errors.New("ровер занят")
		}
		gs.Money -= 50
		rv.Battery = 100
		if err := t.UpdateRover(ctx, rv); err != nil {
			return err
		}
		if err := t.UpdateGameState(ctx, gs); err != nil {
			return err
		}
		return t.AddEvent(ctx, gs.Day, "Батарея заряжена: -50₽")
	})
}

// BuyRover — купить новый ровер.
func (s *Service) BuyRover(ctx context.Context) error {
	const cost = 500
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameStateForUpdate(ctx)
		if err != nil {
			return err
		}
		if gs.Money < cost {
			return errors.New("недостаточно средств")
		}
		gs.Money -= cost
		newID, err := t.CreateRover(ctx, domain.Rover{
			Name:     "Ровер", // имя обновим по фактическому id ниже
			Battery:  100,
			Capacity: 200,
			Speed:    10,
			Status:   "idle",
			X:        BaseX,
			Y:        BaseY,
		})
		if err != nil {
			return err
		}
		if err := t.UpdateRoverName(ctx, newID, "Ровер-"+itoa(newID)); err != nil {
			return err
		}
		if err := t.UpdateGameState(ctx, gs); err != nil {
			return err
		}
		return t.AddEvent(ctx, gs.Day, "Куплен новый ровер: -500₽")
	})
}
