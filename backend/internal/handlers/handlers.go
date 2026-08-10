package handlers

import (
	"encoding/json"
	"net/http"

	"MoonCrisis/internal/repository"
	"MoonCrisis/internal/service"
)

type Handlers struct {
	repo    *repository.Repository
	service *service.Service
}

func New(repo *repository.Repository, svc *service.Service) *Handlers {
	return &Handlers{repo: repo, service: svc}
}

func (h *Handlers) json(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *Handlers) error(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
