// Rider application logic
let map;
let ws;
let currentRide = null;
let markers = {};

// Initialize map
function initMap() {
    // Default to Bangalore coordinates
    map = L.map('map').setView([12.9716, 77.5946], 12);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
    }).addTo(map);

    // Add click event to set pickup/dropoff
    map.on('click', function(e) {
        const lat = e.latlng.lat.toFixed(6);
        const lng = e.latlng.lng.toFixed(6);

        // Toggle between setting pickup and dropoff
        const pickupLat = document.getElementById('pickup-lat');
        const pickupLng = document.getElementById('pickup-lng');

        if (!pickupLat.value || pickupLat.value === '12.9716') {
            pickupLat.value = lat;
            pickupLng.value = lng;
            addMarker('pickup', lat, lng, 'Pickup');
        } else {
            document.getElementById('dropoff-lat').value = lat;
            document.getElementById('dropoff-lng').value = lng;
            addMarker('dropoff', lat, lng, 'Dropoff');
        }
    });
}

// Add marker to map
function addMarker(id, lat, lng, label) {
    if (markers[id]) {
        map.removeLayer(markers[id]);
    }

    const icon = L.divIcon({
        className: 'custom-marker',
        html: `<div style="background: ${id === 'pickup' ? '#3498db' : '#e74c3c'}; color: white; padding: 5px 10px; border-radius: 4px; font-weight: bold;">${label}</div>`
    });

    markers[id] = L.marker([lat, lng], { icon }).addTo(map);
}

// Fetch random rider
async function fetchRandomRider() {
    console.log('[Rider] Fetching random rider...');
    try {
        const response = await fetch('/v1/riders/random');
        console.log('[Rider] Response status:', response.status);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        console.log('[Rider] Received data:', data);

        if (data.id) {
            document.getElementById('rider-id').value = data.id;
            console.log('[Rider] ✓ Rider ID set to:', data.id);
            showNotification('Welcome ' + (data.name || 'Rider'), 'success');
        } else {
            console.warn('[Rider] No ID in response, using fallback');
            document.getElementById('rider-id').value = 'e947a893-6233-4c3c-bdf1-da1742543cbe';
            showNotification('Using default rider ID', 'info');
        }
    } catch (error) {
        console.error('[Rider] ✗ Error fetching random rider:', error);
        // Use default fallback if API fails
        const fallbackID = 'e947a893-6233-4c3c-bdf1-da1742543cbe';
        document.getElementById('rider-id').value = fallbackID;
        console.log('[Rider] Using fallback ID:', fallbackID);
        showNotification('Using default rider ID', 'info');
    }
}

// Connect to WebSocket
function connectWebSocket() {
    const riderID = document.getElementById('rider-id').value;
    console.log('[Rider] Connecting WebSocket with rider ID:', riderID);

    if (!riderID) {
        console.error('[Rider] ✗ WARNING: Rider ID is empty! WebSocket connection will fail.');
    }

    ws = new WebSocket(`ws://localhost:8080/v1/ws?user_id=${riderID}&user_type=rider`);

    ws.onopen = function() {
        console.log('WebSocket connected');
        showNotification('Connected to server', 'success');
    };

    ws.onmessage = function(event) {
        const message = JSON.parse(event.data);
        handleWebSocketMessage(message);
    };

    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
        showNotification('Connection error', 'error');
    };

    ws.onclose = function() {
        console.log('WebSocket closed');
        showNotification('Disconnected from server', 'error');

        // Reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000);
    };
}

// Handle WebSocket messages
function handleWebSocketMessage(message) {
    console.log('Received message:', message);

    switch(message.type) {
        case 'ride_updated':
            updateRideStatus(message.data);
            break;
        case 'driver_assigned':
            handleDriverAssigned(message.data);
            break;
        case 'ride_accepted':
            handleRideAccepted(message.data);
            break;
        case 'trip_started':
            handleTripStarted(message.data);
            break;
        case 'trip_completed':
            handleTripCompleted(message.data);
            break;
        default:
            console.log('Unknown message type:', message.type);
    }
}

// Request a ride
document.getElementById('request-ride').addEventListener('click', function() {
    const riderID = document.getElementById('rider-id').value;

    // Validate rider ID before making request
    if (!riderID) {
        console.error('[Rider] ✗ Cannot request ride: Rider ID is empty!');
        showNotification('Error: Rider ID not set. Please refresh the page.', 'error');
        return;
    }

    console.log('[Rider] Requesting ride with rider ID:', riderID);

    const pickupLat = parseFloat(document.getElementById('pickup-lat').value);
    const pickupLng = parseFloat(document.getElementById('pickup-lng').value);
    const dropoffLat = parseFloat(document.getElementById('dropoff-lat').value);
    const dropoffLng = parseFloat(document.getElementById('dropoff-lng').value);
    const vehicleType = document.getElementById('vehicle-type').value;

    const rideRequest = {
        rider_id: riderID,
        pickup_latitude: pickupLat,
        pickup_longitude: pickupLng,
        dropoff_latitude: dropoffLat,
        dropoff_longitude: dropoffLng,
        vehicle_type: vehicleType
    };

    fetch('/v1/rides', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Idempotency-Key': generateUUID()
        },
        body: JSON.stringify(rideRequest)
    })
    .then(response => response.json())
    .then(data => {
        console.log('Ride requested:', data);
        showRideStatus(data);
        showNotification('Ride requested successfully!', 'success');
    })
    .catch(error => {
        console.error('Error requesting ride:', error);
        showNotification('Failed to request ride', 'error');
    });
});

// Show ride status
function showRideStatus(ride) {
    document.getElementById('ride-status').style.display = 'block';
    document.getElementById('ride-id').textContent = ride.id || 'Generated';

    currentRide = ride;

    // Check if driver is already assigned
    if (ride.status === 'assigned' && ride.driver) {
        document.getElementById('status-text').textContent = 'Driver Assigned ✓';
        document.getElementById('driver-name').textContent = ride.driver_name || ride.driver.name || 'Unknown Driver';
        document.getElementById('eta').textContent = ride.estimated_arrival || '5 mins';
        document.getElementById('fare').textContent = ride.estimated_fare || '-';

        // Add driver marker to map
        if (ride.driver.latitude && ride.driver.longitude) {
            addDriverMarker(ride.driver.latitude, ride.driver.longitude);
        }

        showNotification('Driver found: ' + (ride.driver_name || 'Driver'), 'success');
    } else {
        document.getElementById('status-text').textContent = 'Searching for driver...';
        document.getElementById('driver-name').textContent = 'Searching...';
    }

    // Add markers
    addMarker('pickup', ride.pickup_latitude || document.getElementById('pickup-lat').value,
              ride.pickup_longitude || document.getElementById('pickup-lng').value, 'Pickup');
    addMarker('dropoff', ride.dropoff_latitude || document.getElementById('dropoff-lat').value,
              ride.dropoff_longitude || document.getElementById('dropoff-lng').value, 'Dropoff');
}

// Add driver marker to map
function addDriverMarker(lat, lng) {
    if (markers['driver']) {
        map.removeLayer(markers['driver']);
    }

    const icon = L.divIcon({
        className: 'driver-marker',
        html: '<div style="background: #27ae60; color: white; padding: 8px 12px; border-radius: 4px; font-weight: bold;">🚗 Driver</div>'
    });

    markers['driver'] = L.marker([lat, lng], { icon }).addTo(map);

    // Center map to show all markers
    const bounds = L.latLngBounds([
        [document.getElementById('pickup-lat').value, document.getElementById('pickup-lng').value],
        [lat, lng]
    ]);
    map.fitBounds(bounds, { padding: [50, 50] });
}

// Update ride status
function updateRideStatus(data) {
    document.getElementById('status-text').textContent = data.status || 'Unknown';
}

// Handle driver assigned
function handleDriverAssigned(data) {
    document.getElementById('driver-name').textContent = data.driver_name || 'Driver assigned';
    document.getElementById('eta').textContent = data.eta || '5 mins';
    showNotification('Driver assigned!', 'success');
}

// Handle ride accepted by driver
function handleRideAccepted(data) {
    console.log('Ride accepted by driver:', data);

    document.getElementById('status-text').textContent = 'Driver Accepted ✓';
    document.getElementById('driver-name').textContent = 'Driver on the way';
    document.getElementById('eta').textContent = data.eta || '5 mins';

    showNotification(data.message || 'Driver accepted! They are on the way.', 'success');
}

// Handle trip started
function handleTripStarted(data) {
    console.log('Trip started:', data);
    document.getElementById('status-text').textContent = 'Trip In Progress 🚗';
    showNotification('Trip has started!', 'success');
}

// Handle trip completed
function handleTripCompleted(data) {
    console.log('Trip completed:', data);
    document.getElementById('status-text').textContent = 'Trip Completed ✓';
    if (data.fare) {
        document.getElementById('fare').textContent = data.fare;
    }
    showNotification('Trip completed! Fare: $' + (data.fare || '0.00'), 'success');
}

// Show notification
function showNotification(message, type = 'success') {
    const container = document.getElementById('notifications');
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.textContent = message;
    container.appendChild(notification);

    setTimeout(() => {
        notification.remove();
    }, 5000);
}

// Generate UUID
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        const r = Math.random() * 16 | 0;
        const v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

// Initialize on page load
window.addEventListener('DOMContentLoaded', async function() {
    initMap();
    await fetchRandomRider(); // Auto-fetch a rider
    connectWebSocket();
});
