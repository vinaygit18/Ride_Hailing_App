-- Drop trigger
DROP TRIGGER IF EXISTS update_trips_updated_at ON trips;

-- Drop indexes
DROP INDEX IF EXISTS idx_trips_active;
DROP INDEX IF EXISTS idx_trips_created_at;
DROP INDEX IF EXISTS idx_trips_status;
DROP INDEX IF EXISTS idx_trips_ride_id;

-- Drop table
DROP TABLE IF EXISTS trips;

-- Drop custom type
DROP TYPE IF EXISTS trip_status;
