-- Drop trigger
DROP TRIGGER IF EXISTS update_rides_updated_at ON rides;

-- Drop indexes
DROP INDEX IF EXISTS idx_rides_active;
DROP INDEX IF EXISTS idx_rides_status_created;
DROP INDEX IF EXISTS idx_rides_driver_status;
DROP INDEX IF EXISTS idx_rides_rider_status;
DROP INDEX IF EXISTS idx_rides_idempotency_key;
DROP INDEX IF EXISTS idx_rides_status;
DROP INDEX IF EXISTS idx_rides_driver_id;
DROP INDEX IF EXISTS idx_rides_rider_id;

-- Drop table
DROP TABLE IF EXISTS rides;

-- Drop custom type
DROP TYPE IF EXISTS ride_status;
