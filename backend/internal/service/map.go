package service

import (
	"math"

	"MoonCrisis/internal/domain"
)

// Карта 10x10, база внизу посередине
const (
	MapSize = 10
	BaseX   = 4 // база занимает клетки (4,9) и (5,9)
	BaseY   = 9
)

// zone — характеристики зоны карты
type zone struct {
	speedMult float64
	risk      int
}

// zoneAt возвращает зону по координате.
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

// batteryPerCell — расход батареи на клетку.
func batteryPerCell(weight int) float64 {
	return 2 + float64(weight)/100
}

// travelDays — время на путь в одну сторону от базы до точки.
// dist — манхэттенское расстояние; вес замедляет движение (дольше туда).
func travelDays(rv domain.Rover, x, y int, weight int) int {
	dist := math.Abs(float64(x-BaseX)) + math.Abs(float64(y-BaseY))
	z := zoneAt(x, y)
	days := dist / (float64(rv.Speed) * z.speedMult)
	weightFactor := 1.0 + float64(weight)/600.0 // ~+25% на 150 кг
	return int(math.Ceil(days * weightFactor))
}

// batteryCost — расход батареи на весь маршрут (туда и обратно).
func batteryCost(x, y int, weight int) float64 {
	dist := math.Abs(float64(x-BaseX)) + math.Abs(float64(y-BaseY))
	return dist*batteryPerCell(weight) + dist*batteryPerCell(0)
}

// roverPos — позиция ровера на маршруте «база → заказ → база» на шаге step.
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
