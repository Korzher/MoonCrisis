package service

import (
	"math"
)

// Карта 10x10, база внизу посередине
const (
	BaseX = 4 // база занимает клетки (4,9) и (5,9)
	BaseY = 9
)

// zone — характеристики зоны карты
type zone struct {
	risk int
}

// zoneAt возвращает зону по координате.
func zoneAt(x, y int) zone {
	switch {
	case y >= 7 && x > 0 && x < 9:
		return zone{risk: 0}
	case y >= 4:
		return zone{risk: 20}
	default:
		return zone{risk: 40}
	}
}

// batteryPerCell — расход батареи на клетку.
func batteryPerCell(weight int) float64 {
	return 2 + float64(weight)/100
}

// batteryCost — расход батареи на весь маршрут (туда и обратно).
func batteryCost(x, y int, weight int) float64 {
	dist := math.Abs(float64(x-BaseX)) + math.Abs(float64(y-BaseY))
	return dist*batteryPerCell(weight) + dist*batteryPerCell(0)
}

// moveToward — двигает ровер на шаг в день к цели на rv.Speed клеток в день.
func moveToward(x, y, tx, ty, speed int) (int, int) {
	for i := 0; i < speed && (x != tx || y != ty); i++ {
		if x != tx {
			if x < tx {
				x++
			} else {
				x--
			}
		} else if y != ty {
			if y < ty {
				y++
			} else {
				y--
			}
		}
	}
	return x, y
}
