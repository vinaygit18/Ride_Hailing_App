// Driver Dashboard - Centralized Control Panel
let map;
let ws;
let pendingRequests = {}; // Map of driverID -> rideRequest

// Initialize map
function initMap() {
    map = L.map('map').setView([12.9716, 77.5946], 12);

    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© OpenStreetMap contributors'
    }).addTo(map);
}

// Connect to WebSocket as Dashboard
function connectWebSocket() {
    // Connect as "dashboard" user type to receive all driver notifications
    ws = new WebSocket(`ws://localhost:8080/v1/ws?user_id=dashboard&user_type=dashboard`);

    ws.onopen = function() {
        console.log('[Dashboard] WebSocket connected');
        showNotification('Dashboard connected to server', 'success');
    };

    ws.onmessage = function(event) {
        const message = JSON.parse(event.data);
        handleWebSocketMessage(message);
    };

    ws.onerror = function(error) {
        console.error('[Dashboard] WebSocket error:', error);
    };

    ws.onclose = function() {
        console.log('[Dashboard] WebSocket closed');
        showNotification('Dashboard disconnected. Reconnecting...', 'error');
        setTimeout(connectWebSocket, 3000);
    };
}

// Handle WebSocket messages
function handleWebSocketMessage(message) {
    console.log('[Dashboard] Received message:', message);

    switch(message.type) {
        case 'ride_request':
            handleRideRequest(message.data);
            break;
        case 'ride_accepted':
            handleRideAccepted(message.data);
            break;
        case 'ride_cancelled':
            handleRideCancelled(message.data);
            break;
        case 'trip_completed':
            handleTripCompleted(message.data);
            break;
        case 'driver_status_changed':
            fetchAllDrivers(); // Refresh list
            break;
        default:
            console.log('[Dashboard] Unknown message type:', message.type);
    }
}

// Handle incoming ride request
function handleRideRequest(data) {
    console.log('[Dashboard] New ride request for driver:', data.driver_id || 'unknown');

    // Store the pending request
    const driverID = data.driver_id;
    if (driverID) {
        pendingRequests[driverID] = data;
        console.log('[Dashboard] Stored request for driver:', driverID);

        // Refresh the display to show the request
        fetchAllDrivers();

        // Show notification
        showNotification(`🚕 New ride request for driver!`, 'success');

        // Play sound
        playNotificationSound();
    }
}

// Handle ride acceptance
function handleRideAccepted(data) {
    console.log('[Dashboard] Ride accepted:', data);

    // Remove from pending requests
    const driverID = data.driver_id;
    if (driverID && pendingRequests[driverID]) {
        delete pendingRequests[driverID];
        fetchAllDrivers(); // Refresh to remove request UI
    }

    showNotification(`✅ Ride ${data.ride_id} accepted!`, 'success');
}

// Handle ride cancellation
function handleRideCancelled(data) {
    const driverID = data.driver_id;
    if (driverID && pendingRequests[driverID]) {
        delete pendingRequests[driverID];
        fetchAllDrivers();
    }

    showNotification(`❌ Ride request cancelled`, 'info');
}

// Handle trip completion
function handleTripCompleted(data) {
    console.log('[Dashboard] Trip completed:', data);

    // Refresh to show updated earnings and clear current ride
    fetchAllDrivers();

    const fare = data.fare || data.total_fare || 0;
    const driverName = data.driver_name || 'Driver';
    showNotification(`🎉 ${driverName} completed trip! Earned $${fare.toFixed(2)}`, 'success');
    playNotificationSound();
}

// Accept ride from dashboard
function acceptRide(driverID, rideID) {
    console.log('[Dashboard] Accepting ride:', rideID, 'for driver:', driverID);

    fetch(`/v1/drivers/${driverID}/accept`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            ride_id: rideID
        })
    })
    .then(response => response.json())
    .then(data => {
        console.log('[Dashboard] Ride accepted:', data);

        // Remove from pending requests
        delete pendingRequests[driverID];

        // Refresh the table
        fetchAllDrivers();

        showNotification(`✅ Ride accepted for driver!`, 'success');
    })
    .catch(error => {
        console.error('[Dashboard] Error accepting ride:', error);
        showNotification('Failed to accept ride', 'error');
    });
}

// Decline ride from dashboard
function declineRide(driverID, rideID) {
    console.log('[Dashboard] Declining ride:', rideID, 'for driver:', driverID);

    // Remove from pending requests
    delete pendingRequests[driverID];

    // Refresh the table
    fetchAllDrivers();

    showNotification(`❌ Ride declined`, 'info');
}

// End trip from dashboard
function endTrip(driverID, rideID) {
    console.log('[Dashboard] Ending trip:', rideID, 'for driver:', driverID);

    // Calculate random distance and duration for demo
    const distanceKm = (Math.random() * 20 + 5).toFixed(2); // 5-25 km
    const durationMinutes = Math.floor(Math.random() * 40 + 10); // 10-50 minutes

    fetch(`/v1/trips/${rideID}/end`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            driver_id: driverID,
            distance_km: parseFloat(distanceKm),
            duration_minutes: durationMinutes
        })
    })
    .then(response => response.json())
    .then(data => {
        console.log('[Dashboard] Trip ended:', data);

        // Refresh the table to show updated earnings
        fetchAllDrivers();

        const fare = data.fare || data.total_fare || 0;
        showNotification(`✅ Trip completed! Fare: $${fare.toFixed(2)}`, 'success');
    })
    .catch(error => {
        console.error('[Dashboard] Error ending trip:', error);
        showNotification('Failed to end trip', 'error');
    });
}

// Fetch all drivers
async function fetchAllDrivers() {
    try {
        const response = await fetch('/v1/drivers/all');
        const data = await response.json();

        if (data.drivers) {
            displayDrivers(data.drivers);
        }
    } catch (error) {
        console.error('[Dashboard] Error fetching all drivers:', error);
        document.getElementById('drivers-tbody').innerHTML =
            '<tr><td colspan="5" style="padding: 20px; text-align: center; color: #e74c3c;">Failed to load drivers</td></tr>';
    }
}

// Display drivers in table
function displayDrivers(drivers) {
    const tbody = document.getElementById('drivers-tbody');

    // Update counts
    const onlineDrivers = drivers.filter(d => d.status === 'online');
    const offlineDrivers = drivers.filter(d => d.status === 'offline');
    const activeRides = drivers.filter(d => d.current_ride).length;

    document.getElementById('drivers-count').textContent = drivers.length;
    document.getElementById('online-count').textContent = onlineDrivers.length;
    document.getElementById('offline-count').textContent = offlineDrivers.length;
    document.getElementById('active-rides').textContent = activeRides;

    // Sort: drivers with requests first, then online, then by name
    drivers.sort((a, b) => {
        const aHasRequest = pendingRequests[a.id] !== undefined;
        const bHasRequest = pendingRequests[b.id] !== undefined;

        if (aHasRequest && !bHasRequest) return -1;
        if (!aHasRequest && bHasRequest) return 1;
        if (a.status === 'online' && b.status === 'offline') return -1;
        if (a.status === 'offline' && b.status === 'online') return 1;
        return (a.name || '').localeCompare(b.name || '');
    });

    // Build table rows
    if (drivers.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" style="padding: 20px; text-align: center; color: #666;">No drivers found</td></tr>';
        return;
    }

    tbody.innerHTML = drivers.map(driver => {
        const statusColor = driver.status === 'online' ? '#27ae60' : '#e74c3c';
        const statusIcon = driver.status === 'online' ? '●' : '○';

        // Check if driver has a pending request
        const hasPendingRequest = pendingRequests[driver.id];

        // Current status display
        let currentDisplay = '';
        if (hasPendingRequest) {
            currentDisplay = '<span style="color: #ff9800; font-weight: bold; animation: pulse 1s infinite;">📥 NEW REQUEST</span>';
        } else if (driver.current_ride) {
            const shortRideId = driver.current_ride.substring(driver.current_ride.lastIndexOf('-') + 1, driver.current_ride.lastIndexOf('-') + 7);
            currentDisplay = `<span style="color: #ff9800; font-weight: bold;">🚕 On Ride</span><div style="font-size: 9px; color: #666;">ID: ${shortRideId}</div>`;
        } else if (driver.status === 'online') {
            currentDisplay = '<span style="color: #27ae60;">✓ Available</span>';
        } else {
            currentDisplay = '<span style="color: #999;">-</span>';
        }

        // Earnings display
        const earnings = driver.total_earnings || 0;
        const earningsColor = earnings > 0 ? '#27ae60' : '#666';
        const earningsDisplay = earnings > 0 ? `$${earnings.toFixed(2)}` : '$0.00';

        // Action buttons
        let actionButtons = '';
        if (hasPendingRequest) {
            const request = pendingRequests[driver.id];
            const rideID = request.ride_id;
            actionButtons = `
                <button onclick="acceptRide('${driver.id}', '${rideID}')"
                        class="btn btn-primary"
                        style="padding: 4px 10px; font-size: 11px; margin-right: 5px;">
                    ✓ Accept
                </button>
                <button onclick="declineRide('${driver.id}', '${rideID}')"
                        class="btn btn-danger"
                        style="padding: 4px 10px; font-size: 11px;">
                    ✗ Decline
                </button>
            `;
        } else if (driver.current_ride) {
            // Driver is on a ride - show End Trip button
            actionButtons = `
                <button onclick="endTrip('${driver.id}', '${driver.current_ride}')"
                        class="btn btn-success"
                        style="padding: 4px 10px; font-size: 11px; background: #27ae60;">
                    🏁 End Trip
                </button>
            `;
        } else {
            actionButtons = '<span style="color: #999;">-</span>';
        }

        // Row style - highlight if has pending request
        const rowStyle = hasPendingRequest ? 'background: #fff3cd;' : '';

        return `
            <tr style="${rowStyle}">
                <td style="padding: 10px 8px; border-bottom: 1px solid #eee;">
                    <div style="font-weight: bold;">${driver.name || 'Driver ' + driver.id.substring(0, 8)}</div>
                    <div style="font-size: 10px; color: #666;">${driver.vehicle_type || 'economy'} · ⭐ ${parseFloat(driver.rating || 0).toFixed(1)}</div>
                    ${hasPendingRequest ? '<div style="font-size: 10px; color: #ff9800; margin-top: 2px;">📍 ' + pendingRequests[driver.id].pickup_latitude + ', ' + pendingRequests[driver.id].pickup_longitude + '</div>' : ''}
                </td>
                <td style="padding: 10px 8px; text-align: center; border-bottom: 1px solid #eee;">
                    <span style="color: ${statusColor}; font-size: 16px;">${statusIcon}</span>
                    <div style="font-size: 10px; color: ${statusColor};">${driver.status}</div>
                </td>
                <td style="padding: 10px 8px; text-align: center; border-bottom: 1px solid #eee; font-size: 11px;">
                    ${currentDisplay}
                </td>
                <td style="padding: 10px 8px; text-align: center; border-bottom: 1px solid #eee;">
                    <span style="font-size: 12px; color: ${earningsColor}; font-weight: bold;">${earningsDisplay}</span>
                    <div style="font-size: 9px; color: #666;">today</div>
                </td>
                <td style="padding: 10px 8px; text-align: center; border-bottom: 1px solid #eee;">
                    ${actionButtons}
                </td>
            </tr>
        `;
    }).join('');

    console.log(`[Dashboard] Displayed ${drivers.length} drivers (${onlineDrivers.length} online, ${offlineDrivers.length} offline, ${activeRides} on rides)`);
}

// Play notification sound
function playNotificationSound() {
    try {
        const audioContext = new (window.AudioContext || window.webkitAudioContext)();
        const oscillator = audioContext.createOscillator();
        oscillator.type = 'sine';
        oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
        oscillator.connect(audioContext.destination);
        oscillator.start();
        oscillator.stop(audioContext.currentTime + 0.2);
    } catch (error) {
        console.log('[Dashboard] Could not play sound:', error);
    }
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

// Refresh drivers list button
document.getElementById('refresh-drivers').addEventListener('click', function() {
    showNotification('Refreshing drivers list...', 'info');
    fetchAllDrivers();
});

// Initialize on page load
window.addEventListener('DOMContentLoaded', async function() {
    initMap();
    await fetchAllDrivers(); // Fetch all drivers list
    connectWebSocket(); // Connect to dashboard WebSocket

    // Auto-refresh drivers list every 10 seconds
    setInterval(() => {
        fetchAllDrivers();
    }, 10000);
});

// Add CSS animation for pulsing
const style = document.createElement('style');
style.textContent = `
    @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.5; }
    }
`;
document.head.appendChild(style);
