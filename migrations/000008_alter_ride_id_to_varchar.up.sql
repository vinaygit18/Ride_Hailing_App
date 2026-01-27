-- Change ride ID from UUID to VARCHAR to support timestamp-based IDs
ALTER TABLE trips DROP CONSTRAINT trips_ride_id_fkey;
ALTER TABLE rides ALTER COLUMN id TYPE VARCHAR(255);
ALTER TABLE trips ALTER COLUMN ride_id TYPE VARCHAR(255);

-- Re-add foreign key constraint
ALTER TABLE trips ADD CONSTRAINT trips_ride_id_fkey
    FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE;

COMMENT ON COLUMN rides.id IS 'Ride identifier (timestamp-based or UUID)';
