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
	case y >= 7 && x > 0 && x < 9: // Плато у базы (низ, кроме краёв)
		return zone{speedMult: 1.0, risk: 0}
	case y >= 4: // Горная гряда (середина + края низа)
		return zone{speedMult: 0.7, risk: 20}
	default: // Кратерное поле (верх)
		return zone{speedMult: 0.5, risk: 40}
	}
}

// Расход батареи на клетку: базовый 5 + вес/100
func batteryPerCell(weight int) float64 {
	return 5 + float64(weight)/100
}

// Время пути в днях: манхэттенское расстояние / скорость ровера / множитель зоны
func travelDays(rv domain.Rover, x, y int) int {
	dist := math.Abs(float64(x-BaseX)) + math.Abs(float64(y-BaseY))
	z := zoneAt(x, y)
	days := dist / (float64(rv.Speed) * z.speedMult)
	return int(math.Ceil(days))
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
	rv, err := s.repo.GetRover(ctx, roverID)
	if err != nil {
		return err
	}
	o, err := s.repo.GetOrder(ctx, orderID)
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

	// Помечаем ровер и заказ
	rv.Status = "on_mission"
	if err := s.repo.UpdateRover(ctx, rv); err != nil {
		return err
	}
	if err := s.repo.UpdateOrderStatus(ctx, o.ID, "active"); err != nil {
		return err
	}

	// Создаём доставку
	gs, err := s.repo.GetGameState(ctx)
	if err != nil {
		return err
	}
	_, err = s.repo.CreateDelivery(ctx, domain.Delivery{
		RoverID:    rv.ID,
		OrderID:    o.ID,
		StartedDay: gs.Day,
	})
	return err
}

// NextDay — завершить день: обработать доставки, сгенерировать заказы, проверить конец игры
func (s *Service) NextDay(ctx context.Context) error {
	gs, err := s.repo.GetGameState(ctx)
	if err != nil {
		return err
	}
	if gs.GameOver {
		return errors.New("игра окончена")
	}

	gs.Day++

	// Обработка активных доставок
	deliveries, err := s.repo.ListActiveDeliveries(ctx)
	if err != nil {
		return err
	}
	for _, d := range deliveries {
		rv, err := s.repo.GetRover(ctx, d.RoverID)
		if err != nil {
			return err
		}
		o, err := s.repo.GetOrder(ctx, d.OrderID)
		if err != nil {
			return err
		}

		duration := travelDays(rv, o.X, o.Y)
		if gs.Day >= d.StartedDay+duration {
			// Доставка завершена — проверяем риск
			z := zoneAt(o.X, o.Y)
			risk := z.risk + o.Risk
			if rand.Intn(100) < risk {
				// Провал
				gs.Rating -= 10
				rv.Battery = 0
				rv.Status = "broken"
				rv.X, rv.Y = BaseX, BaseY
				s.repo.UpdateRover(ctx, rv)
				s.repo.UpdateOrderStatus(ctx, o.ID, "failed")
				s.repo.CompleteDelivery(ctx, d.ID, gs.Day, "failed", duration)
				s.repo.AddEvent(ctx, gs.Day, "Доставка «"+o.Title+"» провалена: ровер сломан")
			} else {
				// Успех
				gs.Money += o.Reward
				gs.Rating += 5
				rv.Battery -= int(batteryCost(o.X, o.Y, o.Weight))
				if rv.Battery < 0 {
					rv.Battery = 0
				}
				rv.Status = "idle"
				rv.X, rv.Y = BaseX, BaseY
				s.repo.UpdateRover(ctx, rv)
				s.repo.UpdateOrderStatus(ctx, o.ID, "completed")
				s.repo.CompleteDelivery(ctx, d.ID, gs.Day, "success", duration)
				s.repo.AddEvent(ctx, gs.Day, "Доставка «"+o.Title+"» выполнена: +"+itoa(o.Reward)+"₽")
			}
		}
	}

	// Просроченные заказы
	orders, err := s.repo.ListOrders(ctx)
	if err != nil {
		return err
	}
	for _, o := range orders {
		if o.Status == "active" && o.Deadline < gs.Day {
			gs.Rating -= 5
			s.repo.UpdateOrderStatus(ctx, o.ID, "expired")
			s.repo.AddEvent(ctx, gs.Day, "Заказ «"+o.Title+"» просрочен: рейтинг -5")
		}
	}

	// Генерация новых заказов (2-3 в день)
	for i := 0; i < 2+rand.Intn(2); i++ {
		s.generateOrder(ctx, gs.Day)
	}

	// Конец игры
	if gs.Rating <= 0 {
		gs.GameOver = true
		s.repo.AddEvent(ctx, gs.Day, "Рейтинг базы упал до нуля — игра окончена")
	}
	if gs.Day >= 20 {
		gs.GameOver = true
		s.repo.AddEvent(ctx, gs.Day, "Достигнут 20-й день — игра окончена")
	}

	return s.repo.UpdateGameState(ctx, gs)
}

// generateOrder — создать случайный заказ
func (s *Service) generateOrder(ctx context.Context, day int) {
	x := rand.Intn(10)
	y := rand.Intn(10)
	z := zoneAt(x, y)

	weight := 50 + rand.Intn(450) // 50..500 кг
	reward := weight/10 + rand.Intn(50)
	deadline := day + 2 + rand.Intn(5)
	risk := z.risk/2 + rand.Intn(20)

	// Невыполнимый заказ: вес 1000 кг в кратерном поле
	if rand.Intn(20) == 0 {
		weight = 1000
		reward = 500
		risk = 60
		y = rand.Intn(4) // верх карты
	}

	s.repo.CreateOrder(ctx, domain.Order{
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
