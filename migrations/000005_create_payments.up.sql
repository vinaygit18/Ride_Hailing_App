-- Create custom types for payments
CREATE TYPE payment_status AS ENUM ('pending', 'processing', 'completed', 'failed', 'refunded');
CREATE TYPE payment_method AS ENUM ('card', 'wallet', 'cash', 'upi');

-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL UNIQUE REFERENCES trips(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    status payment_status NOT NULL DEFAULT 'pending',
    payment_method payment_method NOT NULL,
    external_transaction_id VARCHAR(255),
    payment_gateway_response JSONB,
    failure_reason TEXT,
    idempotency_key VARCHAR(255) UNIQUE,
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update updated_at on payments table
CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create indexes
CREATE INDEX idx_payments_trip_id ON payments(trip_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_idempotency_key ON payments(idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_payments_external_transaction_id ON payments(external_transaction_id) WHERE external_transaction_id IS NOT NULL;

-- Composite index for status and created_at
CREATE INDEX idx_payments_status_created ON payments(status, created_at DESC);

-- Add comments for documentation
COMMENT ON TABLE payments IS 'Stores payment transaction information';
COMMENT ON COLUMN payments.status IS 'Payment status: pending, processing, completed, failed, refunded';
COMMENT ON COLUMN payments.payment_method IS 'Payment method used: card, wallet, cash, upi';
COMMENT ON COLUMN payments.external_transaction_id IS 'Transaction ID from payment gateway';
COMMENT ON COLUMN payments.idempotency_key IS 'Unique key to prevent duplicate payments';
