BEGIN;

-- Игровое состояние (одна строка)
CREATE TABLE IF NOT EXISTS game_state (
    id          SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    day         INT NOT NULL DEFAULT 1,
    money       INT NOT NULL DEFAULT 0,
    rating      INT NOT NULL DEFAULT 100,
    game_over   BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO game_state (id) VALUES (1) ON CONFLICT DO NOTHING;

-- Роверы
CREATE TABLE IF NOT EXISTS rovers (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    battery     INT NOT NULL DEFAULT 100,
    capacity    INT NOT NULL,
    speed       INT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'idle',
    x           INT NOT NULL,
    y           INT NOT NULL
);

-- Заказы
CREATE TABLE IF NOT EXISTS orders (
    id          SERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    weight      INT NOT NULL,
    reward      INT NOT NULL,
    deadline    INT NOT NULL,
    risk        INT NOT NULL DEFAULT 0,
    x           INT NOT NULL,
    y           INT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'available'
);

-- Доставки
CREATE TABLE IF NOT EXISTS deliveries (
    id          SERIAL PRIMARY KEY,
    rover_id    INT NOT NULL REFERENCES rovers(id),
    order_id    INT NOT NULL REFERENCES orders(id),
    started_day INT NOT NULL,
    finish_day  INT,
    result      TEXT,
    duration    INT
);

CREATE TABLE IF NOT EXISTS events (
    id          SERIAL PRIMARY KEY,
    day         INT NOT NULL,
    message     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

COMMIT;