-- Drop trigger
DROP TRIGGER IF EXISTS update_riders_updated_at ON riders;

-- Drop indexes
DROP INDEX IF EXISTS idx_riders_phone;
DROP INDEX IF EXISTS idx_riders_email;

-- Drop table
DROP TABLE IF EXISTS riders;
