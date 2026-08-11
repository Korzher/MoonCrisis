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

			outDays := travelDays(rv, o.X, o.Y, o.Weight) // туда: с грузом (дольше)
			backDays := travelDays(rv, o.X, o.Y, 0)       // обратно: без груза (быстрее)
			totalDays := outDays + backDays               // полный круг

			elapsed := gs.Day - d.StartedDay
			dist := abs(o.X-BaseX) + abs(o.Y-BaseY)

			// 1) Движение ровера: до точки за outDays, обратно за backDays
			if dist > 0 {
				var reached int
				if outDays > 0 && elapsed < outDays {
					reached = (elapsed * dist) / outDays
				} else if backDays > 0 {
					back := elapsed - outDays
					if back < 0 {
						back = 0
					}
					reached = dist + (back*dist)/backDays
				}
				if reached > 2*dist {
					reached = 2 * dist
				}
				if reached < 0 {
					reached = 0
				}
				nx, ny := roverPos(BaseX, BaseY, o.X, o.Y, reached, dist)
				if nx != rv.X || ny != rv.Y {
					rv.X, rv.Y = nx, ny
					_ = t.UpdateRover(ctx, rv)
				}
			}

			// 2) Доехал до точки — начисляем награду (заказ считается выполненным)
			if !isDelivered(d) && gs.Day >= d.StartedDay+outDays {
				z := zoneAt(o.X, o.Y)
				risk := z.risk + o.Risk
				if rand.Intn(100) < risk {
					// Провал: ровер сломался
					gs.Rating -= 10
					rv.Status = "broken"
					rv.X, rv.Y = BaseX, BaseY
					_ = t.UpdateRover(ctx, rv)
					_ = t.UpdateOrderStatus(ctx, o.ID, "failed")
					_ = t.CompleteDelivery(ctx, d.ID, gs.Day, "failed", totalDays)
					_ = t.AddEvent(ctx, gs.Day, "Доставка «"+o.Title+"» провалена: ровер сломан")
					continue
				}
				if o.Deadline < gs.Day {
					// Опоздали: без награды, но возврат всё равно нужен
					_ = t.UpdateOrderStatus(ctx, o.ID, "expired")
					_ = t.MarkDelivered(ctx, d.ID)
					_ = t.AddEvent(ctx, gs.Day, "Заказ «"+o.Title+"» доставлен слишком поздно: без награды")
				} else {
					gs.Money += o.Reward
					gs.Rating += 5
					_ = t.UpdateOrderStatus(ctx, o.ID, "completed")
					_ = t.MarkDelivered(ctx, d.ID)
					_ = t.AddEvent(ctx, gs.Day, "Доставка «"+o.Title+"» выполнена: +"+itoa(o.Reward)+"₽")
				}
			}

			// 3) Вернулся на базу — ровер снова свободен
			if isDelivered(d) && gs.Day >= d.StartedDay+totalDays {
				rv.Battery -= int(batteryCost(o.X, o.Y, o.Weight))
				if rv.Battery < 0 {
					rv.Battery = 0
				}
				rv.Status = "idle"
				rv.X, rv.Y = BaseX, BaseY
				_ = t.UpdateRover(ctx, rv)
				_ = t.CompleteDelivery(ctx, d.ID, gs.Day, "success", totalDays)
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
				gs.Rating -= 2
				t.UpdateOrderStatus(ctx, o.ID, "expired")
				t.AddEvent(ctx, gs.Day, "Заказ «"+o.Title+"» не взят вовремя: рейтинг -2")
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
			t.AddEvent(ctx, gs.Day, "Достигнут 50-й день — игра окончена")
		}

		return t.UpdateGameState(ctx, gs)
	})
}

func isDelivered(d domain.Delivery) bool { return d.Result == "delivered" }
