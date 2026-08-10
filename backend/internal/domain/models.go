package domain

// GameState — игровое состояние (одна строка)
type GameState struct {
	Day      int  `json:"day"`
	Money    int  `json:"money"`
	Rating   int  `json:"rating"`
	GameOver bool `json:"game_over"`
}

// Rover — ровер
type Rover struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Battery  int    `json:"battery"`  // 0..100
	Capacity int    `json:"capacity"` // кг
	Speed    int    `json:"speed"`    // клеток за день
	Status   string `json:"status"`   // idle | on_mission | broken | charging
	X        int    `json:"x"`
	Y        int    `json:"y"`
}

// Order — заказ
type Order struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Weight   int    `json:"weight"`
	Reward   int    `json:"reward"`
	Deadline int    `json:"deadline"` // день, до которого нужно доставить
	Risk     int    `json:"risk"`     // 0..100
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Status   string `json:"status"` // available | active | completed | failed | expired
}

// Delivery — доставка
type Delivery struct {
	ID         int    `json:"id"`
	RoverID    int    `json:"rover_id"`
	OrderID    int    `json:"order_id"`
	StartedDay int    `json:"started_day"`
	FinishDay  *int   `json:"finish_day,omitempty"`
	Result     string `json:"result,omitempty"` // success | failed | expired
	Duration   *int   `json:"duration,omitempty"`
}

// Event — событие лога
type Event struct {
	ID      int    `json:"id"`
	Day     int    `json:"day"`
	Message string `json:"message"`
}
