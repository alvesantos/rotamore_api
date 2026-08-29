CREATE TABLE IF NOT EXISTS rides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    customer_name VARCHAR(150) NOT NULL,
    customer_phone VARCHAR(30) NOT NULL,
    passengers_count INT NOT NULL DEFAULT 1,
    pickup VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    notes TEXT,
    ride_date DATE NOT NULL,
    ride_time VARCHAR(10) NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'agendada',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rides_user_date ON rides(user_id, ride_date DESC);

