package service

import (
	"context"
	"errors"
	"math/rand"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// NextDay — завершить день: обработать доставки, сгенерировать заказы, проверить конец игры.
func (s *Service) NextDay(ctx context.Context) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameStateForUpdate(ctx)
		if err != nil {
			return err
		}
		if gs.GameOver {
			return errors.New("игра окончена")
		}

		gs.Day++

		// Обработка активных доставок
		deliveries, err := t.ListActiveDeliveries(ctx)
		if err != nil {
			return err
		}
		for _, d := range deliveries {
			rv, err := t.GetRover(ctx, d.RoverID)
			if err != nil {
				return err
			}
			o, err := t.GetOrder(ctx, d.OrderID)
			if err != nil {
				return err
			}

			// 1) Ранний разворот: заказ уже просрочен, а мы ещё не доехали
			if !isDelivered(d) && o.Deadline < gs.Day {
				gs.Rating -= 5
				_ = t.UpdateOrderStatus(ctx, o.ID, "expired")
				_ = t.MarkDelivered(ctx, d.ID)
				_ = t.AddEvent(ctx, gs.Day, "Заказ «"+o.Title+"» просрочен — ровер развернулся и возвращается")
			}

			// 2) Движение: к цели с грузом (медленнее), к базе пустым (полная скорость)
			if !isDelivered(d) {
				speed := rv.Speed - o.Weight/40 // -1 за каждые 40 кг
				if speed < 1 {
					speed = 1
				}
				nx, ny := moveToward(rv.X, rv.Y, o.X, o.Y, speed)
				if nx != rv.X || ny != rv.Y {
					rv.X, rv.Y = nx, ny
					_ = t.UpdateRover(ctx, rv)
				}
			} else {
				nx, ny := moveToward(rv.X, rv.Y, BaseX, BaseY, rv.Speed)
				if nx != rv.X || ny != rv.Y {
					rv.X, rv.Y = nx, ny
					_ = t.UpdateRover(ctx, rv)
				}
			}

			// 3) Доехали до точки (и ещё не развернулись) — выполняем заказ
			if !isDelivered(d) && rv.X == o.X && rv.Y == o.Y {
				if rand.Intn(100) < o.Risk {
					// Провал: ровер сломался
					gs.Rating -= 10
					rv.Status = "broken"
					rv.X, rv.Y = BaseX, BaseY
					_ = t.UpdateRover(ctx, rv)
					_ = t.UpdateOrderStatus(ctx, o.ID, "failed")
					_ = t.CompleteDelivery(ctx, d.ID, gs.Day, "failed", 0)
					_ = t.AddEvent(ctx, gs.Day, "Доставка «"+o.Title+"» провалена: ровер сломан, рейтинг -10")
					continue
				}
				gs.Money += o.Reward
				gs.Rating += 5
				_ = t.UpdateOrderStatus(ctx, o.ID, "completed")
				_ = t.MarkDelivered(ctx, d.ID)
				_ = t.AddEvent(ctx, gs.Day, "Доставка «"+o.Title+"» выполнена: +"+itoa(o.Reward)+"₽, рейтинг +5")
			}

			// 4) Вернулись на базу — ровер снова свободен
			if isDelivered(d) && rv.X == BaseX && rv.Y == BaseY {
				result := "success"
				if o.Status != "completed" {
					result = "failed"
				}
				rv.Battery -= int(batteryCost(o.X, o.Y, o.Weight))
				if rv.Battery < 0 {
					rv.Battery = 0
				}
				rv.Status = "idle"
				_ = t.UpdateRover(ctx, rv)
				_ = t.CompleteDelivery(ctx, d.ID, gs.Day, result, 0)
				_ = t.AddEvent(ctx, gs.Day, "Ровер вернулся на базу")
			}
		}

		// Просроченные заказы
		orders, err := t.ListOrders(ctx)
		if err != nil {
			return err
		}
		for _, o := range orders {
			if o.Deadline < gs.Day && o.Status == "active" {
				gs.Rating -= 5
				t.UpdateOrderStatus(ctx, o.ID, "expired")
				t.AddEvent(ctx, gs.Day, "Заказ «"+o.Title+"» просрочен: рейтинг -5")
			} else if o.Deadline < gs.Day && o.Status == "available" {
				gs.Rating -= 3
				t.UpdateOrderStatus(ctx, o.ID, "expired")
				t.AddEvent(ctx, gs.Day, "Заказ «"+o.Title+"» не взят вовремя: рейтинг -3")
			}
		}

		// Генерация новых заказов
		avail := 0
		for _, o := range orders {
			if o.Status == "available" {
				avail++
			}
		}
		for avail < 5 {
			generateOrder(ctx, t, gs.Day)
			avail++
		}

		// Конец игры
		if gs.Rating <= 0 {
			gs.GameOver = true
			t.AddEvent(ctx, gs.Day, "Рейтинг базы упал до нуля — игра окончена")
		}
		if gs.Day >= 50 {
			gs.GameOver = true
			t.AddEvent(ctx, gs.Day, "Игра завершена: вы победили, заработав "+itoa(gs.Money)+" монет")
		}

		return t.UpdateGameState(ctx, gs)
	})
}

func isDelivered(d domain.Delivery) bool { return d.Result == "delivered" }
