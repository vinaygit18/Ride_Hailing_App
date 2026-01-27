-- Create custom type for trip status
CREATE TYPE trip_status AS ENUM ('in_progress', 'completed', 'cancelled');

-- Create trips table
CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id UUID NOT NULL UNIQUE REFERENCES rides(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    distance_km DECIMAL(8, 2),
    duration_minutes INTEGER,
    base_fare DECIMAL(10, 2) NOT NULL,
    distance_fare DECIMAL(10, 2) DEFAULT 0,
    time_fare DECIMAL(10, 2) DEFAULT 0,
    surge_multiplier DECIMAL(3, 2) DEFAULT 1.00 CHECK (surge_multiplier >= 1.00 AND surge_multiplier <= 5.00),
    total_fare DECIMAL(10, 2),
    status trip_status NOT NULL DEFAULT 'in_progress',
    route_polyline TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update updated_at on trips table
CREATE TRIGGER update_trips_updated_at BEFORE UPDATE ON trips
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes
CREATE INDEX idx_trips_ride_id ON trips(ride_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trips_created_at ON trips(created_at DESC);

-- Partial index for active trips
CREATE INDEX idx_trips_active ON trips(status, created_at DESC)
    WHERE status = 'in_progress';

-- Add comments for documentation
COMMENT ON TABLE trips IS 'Stores trip details including fare and route information';
COMMENT ON COLUMN trips.status IS 'Current trip status: in_progress, completed, cancelled';
COMMENT ON COLUMN trips.surge_multiplier IS 'Surge pricing multiplier (1.00 to 5.00)';
COMMENT ON COLUMN trips.total_fare IS 'Final calculated fare including surge';
COMMENT ON COLUMN trips.route_polyline IS 'Encoded polyline of the route taken';
