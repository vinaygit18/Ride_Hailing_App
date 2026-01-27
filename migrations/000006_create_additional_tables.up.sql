-- Create driver_locations table for historical tracking (optional)
CREATE TABLE IF NOT EXISTS driver_locations (
    id BIGSERIAL PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    accuracy DECIMAL(8, 2),
    speed DECIMAL(8, 2),
    heading DECIMAL(5, 2),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on driver_id and timestamp for efficient queries
CREATE INDEX idx_driver_locations_driver_timestamp ON driver_locations(driver_id, timestamp DESC);

-- Create notifications table for tracking sent notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('driver', 'rider')),
    notification_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    is_read BOOLEAN DEFAULT FALSE,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes on notifications
CREATE INDEX idx_notifications_user ON notifications(user_id, user_type, sent_at DESC);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read, sent_at DESC)
    WHERE is_read = FALSE;

-- Add comments
COMMENT ON TABLE driver_locations IS 'Stores historical driver location data for analytics';
COMMENT ON TABLE notifications IS 'Stores notification history for drivers and riders';
