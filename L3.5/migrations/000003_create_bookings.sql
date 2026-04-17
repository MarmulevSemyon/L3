-- +goose Up
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'confirmed', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_bookings_event_user_active
ON bookings(event_id, user_id)
WHERE status IN ('pending', 'confirmed');

CREATE INDEX IF NOT EXISTS ix_bookings_event_status_created
ON bookings(event_id, status, created_at);

-- +goose Down
DROP INDEX IF EXISTS ix_bookings_event_status_created;
DROP INDEX IF EXISTS ux_bookings_event_user_active;
DROP TABLE IF EXISTS bookings;