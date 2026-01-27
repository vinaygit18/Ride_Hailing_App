-- Create custom type for ride status
CREATE TYPE ride_status AS ENUM ('requested', 'assigned', 'accepted', 'started', 'completed', 'cancelled');

-- Create rides table
CREATE TABLE IF NOT EXISTS rides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rider_id UUID NOT NULL REFERENCES riders(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES drivers(id) ON DELETE SET NULL,
    status ride_status NOT NULL DEFAULT 'requested',
    vehicle_type vehicle_type NOT NULL,
    pickup_latitude DECIMAL(10, 8) NOT NULL,
    pickup_longitude DECIMAL(11, 8) NOT NULL,
    dropoff_latitude DECIMAL(10, 8) NOT NULL,
    dropoff_longitude DECIMAL(11, 8) NOT NULL,
    pickup_address VARCHAR(500),
    dropoff_address VARCHAR(500),
    estimated_fare DECIMAL(10, 2),
    estimated_distance_km DECIMAL(8, 2),
    estimated_duration_minutes INTEGER,
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    assigned_at TIMESTAMP WITH TIME ZONE,
    accepted_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancellation_reason TEXT,
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update updated_at on rides table
CREATE TRIGGER update_rides_updated_at BEFORE UPDATE ON rides
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes for fast lookups
CREATE INDEX idx_rides_rider_id ON rides(rider_id);
CREATE INDEX idx_rides_driver_id ON rides(driver_id);
CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_idempotency_key ON rides(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX idx_rides_rider_status ON rides(rider_id, status, created_at DESC);
CREATE INDEX idx_rides_driver_status ON rides(driver_id, status, created_at DESC);
CREATE INDEX idx_rides_status_created ON rides(status, created_at DESC);

-- Partial index for active rides
CREATE INDEX idx_rides_active ON rides(status, created_at DESC)
    WHERE status IN ('requested', 'assigned', 'accepted', 'started');

-- Add comments for documentation
COMMENT ON TABLE rides IS 'Stores ride request and assignment information';
COMMENT ON COLUMN rides.status IS 'Current ride status: requested, assigned, accepted, started, completed, cancelled';
COMMENT ON COLUMN rides.idempotency_key IS 'Unique key to prevent duplicate ride creation';
COMMENT ON COLUMN rides.estimated_fare IS 'Estimated fare before the ride starts';
