# Quick Test Guide

Step-by-step instructions to test the ride-hailing application.

> **Note:** Ensure you've completed the setup from [README.md](README.md) before testing.

---

## How Driver Matching Works

Driver assignment is **fully automatic**. When a rider requests a ride:

1. System queries Redis GEORADIUS for nearby drivers (within 5km)
2. Filters for available drivers not already on a ride
3. Selects the nearest available driver
4. Returns matched driver details in the response

**No manual driver ID is required** - the system handles matching automatically.

---

## Complete Ride Flow Test

```
RIDER requests → System auto-matches DRIVER → DRIVER accepts → Trip completes → Payment processed
```

### Step 1: Set Up a Driver (Go Online)

Open **http://localhost:8080/driver** in Browser Tab 1:

1. Click **"Go Online"** - this registers the driver's location in Redis
2. Keep browser console open (F12) to see WebSocket notifications

> The driver is now available for matching.

### Step 2: Request a Ride (Rider)

Open **http://localhost:8080/rider** in Browser Tab 2:

Enter coordinates:

| Field | Value |
|-------|-------|
| Pickup Latitude | `12.9716` |
| Pickup Longitude | `77.5946` |
| Dropoff Latitude | `12.2958` |
| Dropoff Longitude | `76.6394` |
| Vehicle Type | `economy` |

Click **"Request Ride"**

**What happens automatically:**
- System finds nearest available driver via Redis GEORADIUS
- Driver is assigned to the ride
- Response includes matched driver details
- Driver receives WebSocket notification

**Expected Response:**
```json
{
  "id": "ride-1234567890",
  "status": "assigned",
  "driver": {
    "id": "auto-matched-driver-id",
    "name": "Driver Name",
    "vehicle_type": "economy",
    "rating": 4.8
  }
}
```

### Step 3: Driver Accepts the Ride

The driver receives a WebSocket notification with ride details. Accept using curl:

```bash
# Use the driver_id from the ride response
curl -X POST http://localhost:8080/v1/drivers/{DRIVER_ID}/accept \
  -H "Content-Type: application/json" \
  -d '{"ride_id": "RIDE_ID"}'
```

> Replace `{DRIVER_ID}` and `RIDE_ID` with values from the ride response.

**Expected:** Driver status becomes "busy", rider receives acceptance notification.

### Step 4: Complete the Trip

```bash
curl -X POST http://localhost:8080/v1/trips/{RIDE_ID}/end \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": "DRIVER_ID",
    "distance_km": 2.5,
    "duration_minutes": 15
  }'
```

**Expected:** Trip completed with fare breakdown, driver returns to "online".

### Step 5: Process Payment

```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: payment-$(date +%s)" \
  -d '{
    "trip_id": "RIDE_ID",
    "payment_method": "card",
    "amount": 92.50
  }'
```

**Payment Methods:** `card`, `wallet`, `cash`, `upi`

---

## API Testing Examples

### Create Ride (Driver Auto-Matched)

```bash
curl -X POST http://localhost:8080/v1/rides \
  -H "Content-Type: application/json" \
  -d '{
    "rider_id": "any-valid-rider-uuid",
    "pickup_latitude": 12.9716,
    "pickup_longitude": 77.5946,
    "dropoff_latitude": 12.2958,
    "dropoff_longitude": 76.6394,
    "vehicle_type": "economy"
  }'
```

**Note:** No `driver_id` in request - matching is automatic!

**Vehicle Types:** `economy`, `premium`, `luxury`

### Get Random Rider ID

```bash
curl http://localhost:8080/v1/riders/random
```

### Get Ride Details

```bash
curl http://localhost:8080/v1/rides/{RIDE_ID}
```

### Update Driver Location (Go Online)

```bash
curl -X POST http://localhost:8080/v1/drivers/{DRIVER_ID}/location \
  -H "Content-Type: application/json" \
  -d '{"latitude": 12.9716, "longitude": 77.5946}'
```

This adds the driver to Redis geo-index and available set.

### Get All Drivers

```bash
curl http://localhost:8080/v1/drivers/all
```

### Accept Ride

```bash
curl -X POST http://localhost:8080/v1/drivers/{DRIVER_ID}/accept \
  -H "Content-Type: application/json" \
  -d '{"ride_id": "RIDE_ID"}'
```

### End Trip

```bash
curl -X POST http://localhost:8080/v1/trips/{RIDE_ID}/end \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": "DRIVER_ID",
    "distance_km": 5.2,
    "duration_minutes": 20
  }'
```

### Process Payment (Idempotent)

```bash
curl -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{
    "trip_id": "RIDE_ID",
    "payment_method": "card",
    "amount": 150.00
  }'
```

---

## WebSocket Testing

### Connection URL

```
ws://localhost:8080/v1/ws?user_id=USER_ID&user_type=TYPE
```

### Using wscat

```bash
# Install
npm install -g wscat

# Connect as driver
wscat -c "ws://localhost:8080/v1/ws?user_id=DRIVER_ID&user_type=driver"

# Connect as rider
wscat -c "ws://localhost:8080/v1/ws?user_id=RIDER_ID&user_type=rider"
```

### Message Types

**ride_request** (Driver receives when matched):
```json
{
  "type": "ride_request",
  "data": {
    "ride_id": "ride-1706253445123456789",
    "rider_id": "rider-uuid",
    "pickup_latitude": 12.9716,
    "pickup_longitude": 77.5946,
    "vehicle_type": "economy",
    "estimated_fare": 92.50
  }
}
```

**ride_accepted** (Rider receives):
```json
{
  "type": "ride_accepted",
  "data": {
    "ride_id": "ride-1706253445123456789",
    "driver_id": "matched-driver-uuid",
    "driver_name": "John Doe",
    "estimated_arrival": "5 mins"
  }
}
```

**trip_completed** (Both receive):
```json
{
  "type": "trip_completed",
  "data": {
    "ride_id": "ride-1706253445123456789",
    "total_fare": 92.50,
    "distance_km": 2.5
  }
}
```

---

## Troubleshooting

### "No Drivers Available"

This means no drivers are within 5km radius or none are online.

```bash
# Check if any drivers are available in Redis
docker exec gocomet-redis redis-cli SMEMBERS drivers:available

# Check driver locations
docker exec gocomet-redis redis-cli ZRANGE drivers:locations 0 -1

# Make a driver available by updating location
curl -X POST http://localhost:8080/v1/drivers/{DRIVER_ID}/location \
  -H "Content-Type: application/json" \
  -d '{"latitude": 12.9716, "longitude": 77.5946}'
```

### Services Not Running

```bash
# Check containers
docker ps

# View logs
docker logs gocomet-postgres
docker logs gocomet-redis

# Restart everything
make docker-down && make docker-up && make run
```

### Database Issues

```bash
# Reset migrations
make migrate-down
make migrate-up
make seed
```

### Verify Driver Availability in Redis

```bash
# Check if driver is in available set
docker exec gocomet-redis redis-cli SISMEMBER drivers:available {DRIVER_ID}
# Returns: 1 (available) or 0 (not available)

# Find nearby drivers (within 5km of pickup point)
docker exec gocomet-redis redis-cli GEORADIUS drivers:locations 77.5946 12.9716 5 km WITHDIST COUNT 5
```

### Verify Database State

```bash
# Connect to PostgreSQL
docker exec -it gocomet-postgres psql -U postgres -d gocomet

# Check online drivers
SELECT COUNT(*) FROM drivers WHERE status = 'online';

# Check recent rides with matched drivers
SELECT id, status, driver_id, created_at FROM rides ORDER BY created_at DESC LIMIT 5;
```

### WebSocket Not Connecting

1. Verify server is running: `curl http://localhost:8080/health`
2. Check browser console for errors
3. Ensure correct `user_id` and `user_type` parameters

### WebSocket Not Receiving Messages

- Driver must be "online" and have location set in Redis
- Verify ride was created successfully (check response for driver details)
- Check both UIs are connected (look for "WebSocket connected" in console)

---

## Matching Configuration

The automatic matching uses these defaults:

| Setting | Value | Description |
|---------|-------|-------------|
| MaxRadiusKM | 5.0 | Search within 5 km radius |
| MaxTimeout | 30 | 30 second timeout |
| MaxCandidates | 10 | Check top 10 nearest drivers |

Drivers are selected by:
1. Proximity (closest first via GEORADIUS sorted ASC)
2. Availability (must be in `drivers:available` set)
3. Not on active ride (no `driver:{id}:current_ride` key)

---

## Pricing Reference

| Vehicle Type | Base Fare | Per KM | Per Minute |
|-------------|-----------|--------|------------|
| Economy | 50 | 10 | 2 |
| Premium | 100 | 15 | 3 |
| Luxury | 200 | 25 | 5 |

**Surge Multiplier:** 1.0x - 3.0x (based on demand)

**Example Calculation (Economy, 5km, 20min, no surge):**
- Base: 50
- Distance: 5 × 10 = 50
- Time: 20 × 2 = 40
- **Total: 140**