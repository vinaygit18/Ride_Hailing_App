-- Drop trigger
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;

-- Drop indexes
DROP INDEX IF EXISTS idx_payments_status_created;
DROP INDEX IF EXISTS idx_payments_external_transaction_id;
DROP INDEX IF EXISTS idx_payments_idempotency_key;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_trip_id;

-- Drop table
DROP TABLE IF EXISTS payments;

-- Drop custom types
DROP TYPE IF EXISTS payment_method;
DROP TYPE IF EXISTS payment_status;
