package service

import (
	"context"
	"math/rand"
	"strconv"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// generateOrder — создать случайный заказ.
func generateOrder(ctx context.Context, r *repository.Repository, day int) {
	x := rand.Intn(10)
	y := rand.Intn(10)
	z := zoneAt(x, y)

	weight := 30 + rand.Intn(121) // 30..150 кг
	reward := weight*2 + rand.Intn(40)
	deadline := day + 3 + rand.Intn(6) // 3..8 дней
	risk := 5 + int(float64(rand.Intn(20))*float64(z.risk)/100.0)

	r.CreateOrder(ctx, domain.Order{
		Title:    "Груз-" + itoa(weight) + "кг",
		Weight:   weight,
		Reward:   reward,
		Deadline: deadline,
		Risk:     risk,
		X:        x,
		Y:        y,
		Status:   "available",
	})
}

// itoa — int в строку (без strconv в каждом месте).
func itoa(n int) string {
	return strconv.Itoa(n)
}
