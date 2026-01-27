-- Create driver_earnings table for tracking daily earnings
CREATE TABLE IF NOT EXISTS driver_earnings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    total_rides INTEGER DEFAULT 0,
    total_earnings DECIMAL(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(driver_id, date)
);

-- Create trigger to update updated_at
CREATE TRIGGER update_driver_earnings_updated_at BEFORE UPDATE ON driver_earnings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes
CREATE INDEX idx_driver_earnings_driver_id ON driver_earnings(driver_id);
CREATE INDEX idx_driver_earnings_date ON driver_earnings(date DESC);
CREATE INDEX idx_driver_earnings_driver_date ON driver_earnings(driver_id, date DESC);

-- Add comments for documentation
COMMENT ON TABLE driver_earnings IS 'Tracks driver earnings aggregated by date';
COMMENT ON COLUMN driver_earnings.date IS 'Date for which earnings are tracked';
COMMENT ON COLUMN driver_earnings.total_rides IS 'Number of completed rides on this date';
COMMENT ON COLUMN driver_earnings.total_earnings IS 'Total earnings for this date';
