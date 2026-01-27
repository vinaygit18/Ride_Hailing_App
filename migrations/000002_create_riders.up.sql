-- Create riders table
CREATE TABLE IF NOT EXISTS riders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL UNIQUE,
    rating DECIMAL(3, 2) DEFAULT 5.00 CHECK (rating >= 0 AND rating <= 5),
    total_rides INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update updated_at on riders table
CREATE TRIGGER update_riders_updated_at BEFORE UPDATE ON riders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create index on email for fast lookup
CREATE INDEX idx_riders_email ON riders(email);

-- Create index on phone for fast lookup
CREATE INDEX idx_riders_phone ON riders(phone);

-- Add comment for documentation
COMMENT ON TABLE riders IS 'Stores rider profile information';
COMMENT ON COLUMN riders.rating IS 'Average rating from 0.00 to 5.00';
