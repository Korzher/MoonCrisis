package service

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"strconv"

	"MoonCrisis/internal/domain"
	"MoonCrisis/internal/repository"
)

// Карта 10x10, база внизу посередине
const (
	MapSize = 10
	BaseX   = 4 // база занимает клетки (4,9) и (5,9)
	BaseY   = 9
)

// Зоны: скорость и риск зависят от позиции
type zone struct {
	speedMult float64
	risk      int
}

func zoneAt(x, y int) zone {
	switch {
	case y >= 7 && x > 0 && x < 9:
		return zone{speedMult: 1.00, risk: 0}
	case y >= 4:
		return zone{speedMult: 0.85, risk: 20}
	default:
		return zone{speedMult: 0.75, risk: 40}
	}
}

// Расход батареи на клетку: базовый 5 + вес/100
func batteryPerCell(weight int) float64 {
	return 2 + float64(weight)/100
}

// Время пути в днях: манхэттенское расстояние / скорость ровера / множитель зоны
func travelDays(rv domain.Rover, x, y int, weight int) int {
	dist := math.Abs(float64(x-BaseX)) + math.Abs(float64(y-BaseY))
	z := zoneAt(x, y)
	days := (2 * dist) / (float64(rv.Speed) * z.speedMult)
	weightFactor := 1.0 + float64(weight)/600.0 // ~+25% на 150 кг
	return int(math.Ceil(days * weightFactor))
}

// Расход батареи на весь маршрут (туда и обратно)
func batteryCost(x, y int, weight int) float64 {
	dist := math.Abs(float64(x-BaseX)) + math.Abs(float64(y-BaseY))
	return dist*batteryPerCell(weight) + dist*batteryPerCell(0)
}

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// AssignRover — назначить ровер на заказ
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
		gs, err := t.GetGameState(ctx)
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

// NextDay — завершить день: обработать доставки, сгенерировать заказы, проверить конец игры
func (s *Service) NextDay(ctx context.Context) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameState(ctx)
		if err != nil {
			return err
		}
		if gs.GameOver {
			return errors.New("игра окончена")
		}

		gs.Day++

		deliveries, err := t.ListActiveDeliveries(ctx)
		if err != nil {
			return err
		}

		// Обработка активных доставок
		for _, d := range deliveries {
			rv, err := t.GetRover(ctx, d.RoverID)
			if err != nil {
				return err
			}
			o, err := t.GetOrder(ctx, d.OrderID)
			if err != nil {
				return err
			}

			outDays := travelDays(rv, o.X, o.Y, o.Weight) // время до точки
			totalDays := 2 * outDays                      // туда + обратно
			elapsed := gs.Day - d.StartedDay
			dist := abs(o.X-BaseX) + abs(o.Y-BaseY)

			// 1) Движение ровера по карте (полный круг: туда, потом обратно)
			if dist > 0 {
				total := totalDays
				if total < 1 {
					total = 1
				}
				reached := (elapsed * 2 * dist) / total
				if reached > 2*dist {
					reached = 2 * dist
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
				// Не взяли вовремя — небольшой штраф за упущенный заказ
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
		if gs.Day >= 100 {
			gs.GameOver = true
			t.AddEvent(ctx, gs.Day, "Достигнут 100-й день — игра окончена")
		}

		return t.UpdateGameState(ctx, gs)
	})
}

// generateOrder — создать случайный заказ
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

// itoa — int в строку (без strconv в каждом месте)
func itoa(n int) string {
	return strconv.Itoa(n)
}

func isDelivered(d domain.Delivery) bool { return d.Result == "delivered" }

// RepairRover — починить сломанный ровер
func (s *Service) RepairRover(ctx context.Context, roverID int) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameState(ctx)
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

// ChargeRover — полная зарядка батареи
func (s *Service) ChargeRover(ctx context.Context, roverID int) error {
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameState(ctx)
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

// BuyRover — купить новый ровер
func (s *Service) BuyRover(ctx context.Context) error {
	const cost = 500
	return s.repo.Tx(ctx, func(t *repository.Repository) error {
		gs, err := t.GetGameState(ctx)
		if err != nil {
			return err
		}
		if gs.Money < cost {
			return errors.New("недостаточно средств")
		}
		gs.Money -= cost
		if _, err := t.CreateRover(ctx, domain.Rover{
			Name:     "Ровер-" + itoa(gs.Day+1),
			Battery:  100,
			Capacity: 200,
			Speed:    10,
			Status:   "idle",
			X:        BaseX,
			Y:        BaseY,
		}); err != nil {
			return err
		}
		if err := t.UpdateGameState(ctx, gs); err != nil {
			return err
		}
		return t.AddEvent(ctx, gs.Day, "Куплен новый ровер: -500₽")
	})
}

// GetAvailableRovers — роверы, доступные для назначения
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
		{Title: "Углеводороды", Weight: 140, Reward: 300, Deadline: 6, Risk: 20, X: 3, Y: 3, Status: "available"},
		{Title: "Кислородные баллоны", Weight: 80, Reward: 200, Deadline: 5, Risk: 10, X: 7, Y: 6, Status: "available"},
		{Title: "Образцы реголита", Weight: 40, Reward: 120, Deadline: 4, Risk: 0, X: 2, Y: 8, Status: "available"},
		{Title: "Научное оборудование", Weight: 150, Reward: 350, Deadline: 8, Risk: 30, X: 6, Y: 2, Status: "available"},
		{Title: "Тяжёлый слиток", Weight: 150, Reward: 380, Deadline: 9, Risk: 40, X: 0, Y: 0, Status: "available"},
	}
	for _, o := range orders {
		if _, err := s.repo.CreateOrder(ctx, o); err != nil {
			return err
		}
	}

	return s.repo.AddEvent(ctx, 1, "Новая игра начата. День 1")
}

// roverPos — позиция ровера на маршруте «база → заказ → база» на шаге step.
// dist — кол-во клеток от базы до цели (манхэттен), всего путь = 2*dist.
func roverPos(bx, by, gx, gy, step, dist int) (int, int) {
	if dist == 0 {
		return bx, by
	}
	sx, sy := 1, 1
	if gx < bx {
		sx = -1
	}
	if gy < by {
		sy = -1
	}
	dx, dy := abs(gx-bx), abs(gy-by)

	if step <= dist {
		// идём к цели: сначала по X, потом по Y
		x, y := bx, by
		if rem := step; rem > dx {
			x = gx
			y = by + sy*(rem-dx)
		} else {
			x = bx + sx*rem
			y = by
		}
		return x, y
	}
	// возврат к базе: сперва обратно по Y, потом по X
	x, y := gx, gy
	back := step - dist
	if rem := back; rem >= dy {
		y = by
		x = gx - sx*(rem-dy)
	} else {
		y = gy - sy*rem
		x = gx
	}
	return x, y
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
