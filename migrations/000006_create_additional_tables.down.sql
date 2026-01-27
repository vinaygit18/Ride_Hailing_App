-- Drop indexes
DROP INDEX IF EXISTS idx_notifications_unread;
DROP INDEX IF EXISTS idx_notifications_user;
DROP INDEX IF EXISTS idx_driver_locations_recent;
DROP INDEX IF EXISTS idx_driver_locations_driver_timestamp;

-- Drop tables
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS driver_locations;
