package service

import (
	"context"

	"MoonCrisis/internal/domain"
)

// InitGame — начать новую игру (сброс данных и стартовые сущности).
func (s *Service) InitGame(ctx context.Context) error {
	gs := domain.GameState{Day: 1, Money: 200, Rating: 100, GameOver: false}

	if err := s.repo.ResetGame(ctx); err != nil {
		return err
	}
	if err := s.repo.UpsertGameState(ctx, gs); err != nil {
		return err
	}
	if err := s.repo.UpdateGameState(ctx, gs); err != nil {
		return err
	}

	// Стартовый ровер
	rv := domain.Rover{
		Name:     "Ровер-1",
		Battery:  100,
		Capacity: 100,
		Speed:    8,
		Status:   "idle",
		X:        BaseX,
		Y:        BaseY,
	}
	if _, err := s.repo.CreateRover(ctx, rv); err != nil {
		return err
	}

	// Пять начальных заказов
	orders := []domain.Order{
		{Title: "Углеводороды", Weight: 140, Reward: 300, Deadline: 6, Risk: 7, X: 3, Y: 3, Status: "available"},
		{Title: "Кислородные баллоны", Weight: 80, Reward: 200, Deadline: 5, Risk: 4, X: 7, Y: 6, Status: "available"},
		{Title: "Образцы реголита", Weight: 40, Reward: 120, Deadline: 4, Risk: 0, X: 2, Y: 8, Status: "available"},
		{Title: "Научное оборудование", Weight: 150, Reward: 350, Deadline: 8, Risk: 11, X: 6, Y: 2, Status: "available"},
		{Title: "Тяжёлый слиток", Weight: 150, Reward: 380, Deadline: 9, Risk: 14, X: 0, Y: 0, Status: "available"},
	}
	for _, o := range orders {
		if _, err := s.repo.CreateOrder(ctx, o); err != nil {
			return err
		}
	}

	return s.repo.AddEvent(ctx, 1, "Новая игра начата. День 1")
}
