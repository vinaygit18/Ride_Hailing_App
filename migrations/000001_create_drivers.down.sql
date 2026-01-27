-- Drop triggers
DROP TRIGGER IF EXISTS update_drivers_updated_at ON drivers;

-- Drop indexes
DROP INDEX IF EXISTS idx_drivers_status_vehicle;
DROP INDEX IF EXISTS idx_drivers_vehicle_type;
DROP INDEX IF EXISTS idx_drivers_status;

-- Drop table
DROP TABLE IF EXISTS drivers;

-- Drop custom types
DROP TYPE IF EXISTS vehicle_type;
DROP TYPE IF EXISTS driver_status;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();
