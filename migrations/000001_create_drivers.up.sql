-- Create custom types for drivers
CREATE TYPE driver_status AS ENUM ('online', 'offline', 'busy');
CREATE TYPE vehicle_type AS ENUM ('economy', 'premium', 'luxury');

-- Create drivers table
CREATE TABLE IF NOT EXISTS drivers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL UNIQUE,
    status driver_status NOT NULL DEFAULT 'offline',
    vehicle_type vehicle_type NOT NULL DEFAULT 'economy',
    current_latitude DECIMAL(10, 8),
    current_longitude DECIMAL(11, 8),
    rating DECIMAL(3, 2) DEFAULT 5.00 CHECK (rating >= 0 AND rating <= 5),
    total_rides INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on status for fast lookup of available drivers
CREATE INDEX idx_drivers_status ON drivers(status) WHERE status = 'online';

-- Create index on vehicle_type for filtering
CREATE INDEX idx_drivers_vehicle_type ON drivers(vehicle_type);

-- Create composite index for status and vehicle_type
CREATE INDEX idx_drivers_status_vehicle ON drivers(status, vehicle_type);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update updated_at on drivers table
CREATE TRIGGER update_drivers_updated_at BEFORE UPDATE ON drivers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add comment for documentation
COMMENT ON TABLE drivers IS 'Stores driver profile information and current status';
COMMENT ON COLUMN drivers.status IS 'Current availability status: online (available), offline (not available), busy (on a ride)';
COMMENT ON COLUMN drivers.vehicle_type IS 'Type of vehicle: economy, premium, or luxury';
COMMENT ON COLUMN drivers.rating IS 'Average rating from 0.00 to 5.00';
