package service

import (
	"context"

	"MoonCrisis/internal/domain"
)

// GetAvailableRovers — роверы, доступные для назначения.
func (s *Service) GetAvailableRovers(ctx context.Context) ([]domain.Rover, error) {
	rovers, err := s.repo.ListRovers(ctx)
	if err != nil {
		return nil, err
	}
	var avail []domain.Rover
	for _, rv := range rovers {
		if rv.Status == "idle" {
			avail = append(avail, rv)
		}
	}
	return avail, nil
}
