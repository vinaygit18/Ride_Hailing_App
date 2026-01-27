-- Revert ride ID back to UUID (only works if no data or all IDs are valid UUIDs)
ALTER TABLE trips DROP CONSTRAINT trips_ride_id_fkey;
ALTER TABLE rides ALTER COLUMN id TYPE UUID USING id::UUID;
ALTER TABLE trips ALTER COLUMN ride_id TYPE UUID USING ride_id::UUID;

ALTER TABLE trips ADD CONSTRAINT trips_ride_id_fkey
    FOREIGN KEY (ride_id) REFERENCES rides(id) ON DELETE CASCADE;
