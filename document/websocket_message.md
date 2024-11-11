# WebSocket Message Handling for Ride-Sharing Flutter App

This document outlines the different types of WebSocket messages that the Flutter app should handle for the ride-sharing application. Each message type corresponds to a specific action or event in the ride-sharing process.

## Message Types and Structures

### Test websocket

Call to endpoint ```/notification/create-test-websocket```

```json
{
  "type":"test",
  "data": {
    "message": "string"
  }
}
```

### 1. new-give-ride-request

Sent when a driver offers a ride to a hitchhiker.

```json
{
  "type": "new-give-ride-request",
  "data": {
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
    "user": {
      "id": "UUID",
      "phoneNumber": "string",
      "fullName": "string"
    },
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    },
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "start_address": "string",
    "end_address": "string",
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "driver_current_latitude": 0.0,
    "driver_current_longitude": 0.0,
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string",
    "status": "string",
    "fare": 0.0
  }
}
```

### 2. new-hitch-ride-request

Sent when a hitchhiker requests a ride from a driver.

```json
{
  "type": "new-hitch-ride-request",
  "data": {
    "ride_request_id": "UUID",
    "ride_offer_id": "UUID",
    "receiver_id": "UUID",
    "user": {
      "id": "UUID",
      "phoneNumber": "string",
      "fullName": "string"
    },
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    },
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "rider_current_latitude": 0.0,
    "rider_current_longitude": 0.0,
    "start_address": "string",
    "end_address": "string",
    "status": "string",
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string"
  }
}
```

### 3. accept-give-ride-request

Sent when a hitchhiker accepts a ride offer from a driver.

```json
{
  "type": "accept-give-ride-request",
  "data": {
    "ride_id": "UUID",
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
    "status": "string",
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string",
    "start_address": "string",
    "end_address": "string",
    "fare": 0.0,
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "transaction": {
      "transaction_id": "UUID",
      "amount": 0.0,
      "status": "string",
      "payment_method": "string"
    },
    "driver_current_latitude": 0.0,
    "driver_current_longitude": 0.0,
    "rider_current_latitude": 0.0,
    "rider_current_longitude": 0.0,
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    }
  }
}
```

### 4. accept-hitch-ride-request

Sent when a driver accepts a ride request from a hitchhiker.

```json
{
  "type": "accept-hitch-ride-request",
  "data": {
    "ride_id": "UUID",
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
    "status": "string",
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string",
    "start_address": "string",
    "end_address": "string",
    "fare": 0.0,
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "transaction": {
      "transaction_id": "UUID",
      "amount": 0.0,
      "status": "string",
      "payment_method": "string"
    },
    "driver_current_latitude": 0.0,
    "driver_current_longitude": 0.0,
    "rider_current_latitude": 0.0,
    "rider_current_longitude": 0.0,
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    }
  }
}
```

### 5. cancel-give-ride-request

Sent when a hitchhiker cancels a ride offer from a driver.

```json
{
  "type": "cancel-give-ride-request",
  "data": {
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "user_id": "UUID",
    "receiver_id": "UUID"
  }
}
```

### 6. cancel-hitch-ride-request

Sent when a driver cancels a ride request from a hitchhiker.

```json
{
  "type": "cancel-hitch-ride-request",
  "data": {
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "user_id": "UUID",
    "receiver_id": "UUID"
  }
}
```

### 7. start-ride

Send when driver start the ride

```json
{
  "type": "start-ride",
  "data": {
    "ride_id": "UUID",
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
    "status": "string",
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string",
    "start_address": "string",
    "end_address": "string",
    "fare": 0.0,
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "transaction": {
      "transaction_id": "UUID",
      "amount": 0.0,
      "status": "string",
      "payment_method": "string"
    },
    "driver_current_latitude": 0.0,
    "driver_current_longitude": 0.0,
    "rider_current_latitude": 0.0,
    "rider_current_longitude": 0.0,
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    }
  }
}
```

### 8. end-ride

Send when driver end the ride

```json
{
  "type": "end-ride",
  "data": {
    "ride_id": "UUID",
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
    "status": "string",
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string",
    "start_address": "string",
    "end_address": "string",
    "fare": 0.0,
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "transaction": {
      "transaction_id": "UUID",
      "amount": 0.0,
      "status": "string",
      "payment_method": "string"
    },
    "driver_current_latitude": 0.0,
    "driver_current_longitude": 0.0,
    "rider_current_latitude": 0.0,
    "rider_current_longitude": 0.0,
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    }
  }
}
```

### 9. update-ride-location

Send when driver update the ride location

```json
{
  "type": "update-ride-location",
  "data": {
    "ride_id": "UUID",
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
    "status": "string",
    "start_time": "ISO8601 string",
    "end_time": "ISO8601 string",
    "start_address": "string",
    "end_address": "string",
    "fare": 0.0,
    "encoded_polyline": "string",
    "distance": 0.0,
    "duration": 0,
    "transaction": {
      "transaction_id": "UUID",
      "amount": 0.0,
      "status": "string",
      "payment_method": "string"
    },
    "driver_current_latitude": 0.0,
    "driver_current_longitude": 0.0,
    "rider_current_latitude": 0.0,
    "rider_current_longitude": 0.0,
    "start_latitude": 0.0,
    "start_longitude": 0.0,
    "end_latitude": 0.0,
    "end_longitude": 0.0,
    "vehicle": {
      "vehicle_id": "UUID",
      "name": "string",
      "fuel_consumed": 0.0,
      "license_plate": "string"
    }
  }
}
```

### 10. cancel-ride

Send when the driver or hitcher want to cancel the ride

```json
```json
{
  "type": "cancel-ride",
  "data": {
    "ride_id": "UUID",
    "ride_offer_id": "UUID",
    "ride_request_id": "UUID",
    "receiver_id": "UUID",
  }
}
```

## Implementing WebSocket Handling in Flutter

To handle these WebSocket messages in your Flutter application:

1. Establish a WebSocket connection when the user logs in.
2. Listen for incoming messages and parse the JSON payload.
3. Based on the `type` field in the message, dispatch the appropriate action or update the UI.

Here's an example of how you might structure your WebSocket handling in Flutter:

```dart
import 'package:web_socket_channel/web_socket_channel.dart';
import 'dart:convert';
import 'package:flutter/material.dart';

class WebSocketService {
  late WebSocketChannel _channel;
  final String _wsUrl = 'ws://your-backend-url/ws';

  void connect() {
    _channel = WebSocketChannel.connect(Uri.parse(_wsUrl));
    _channel.stream.listen(_handleMessage, onError: _handleError, onDone: _handleDone);
  }

  void _handleMessage(dynamic message) {
    final parsedMessage = jsonDecode(message);
    switch (parsedMessage['type']) {
      case 'new-give-ride-request':
        _handleNewGiveRideRequest(parsedMessage['data']);
        break;
      case 'new-hitch-ride-request':
        _handleNewHitchRideRequest(parsedMessage['data']);
        break;
      case 'accepted-give-ride-request':
        _handleAcceptedGiveRideRequest(parsedMessage['data']);
        break;
      case 'accepted-hitch-ride-request':
        _handleAcceptedHitchRideRequest(parsedMessage['data']);
        break;
      case 'cancel-give-ride-request':
        _handleCancelGiveRideRequest(parsedMessage['data']);
        break;
      case 'cancel-hitch-ride-request':
        _handleCancelHitchRideRequest(parsedMessage['data']);
        break;
      default:
        print('Unknown message type: ${parsedMessage['type']}');
    }
  }

  void _handleNewGiveRideRequest(Map<String, dynamic> data) {
    // Display notification with ride offer details
    // You might want to use a state management solution or event bus to update the UI
    print('New give ride request: $data');
  }

  void _handleNewHitchRideRequest(Map<String, dynamic> data) {
    // Display notification with ride request details
    print('New hitch ride request: $data');
  }

  void _handleAcceptedGiveRideRequest(Map<String, dynamic> data) {
    // Update UI to show accepted ride offer
    print('Accepted give ride request: $data');
  }

  void _handleAcceptedHitchRideRequest(Map<String, dynamic> data) {
    // Update UI to show accepted ride request
    print('Accepted hitch ride request: $data');
  }

  void _handleCancelGiveRideRequest(Map<String, dynamic> data) {
    // Update UI to show cancelled ride offer
    print('Cancelled give ride request: $data');
  }

  void _handleCancelHitchRideRequest(Map<String, dynamic> data) {
    // Update UI to show cancelled ride request
    print('Cancelled hitch ride request: $data');
  }

  void _handleError(error) {
    print('WebSocket error: $error');
    // Implement reconnection logic here
  }

  void _handleDone() {
    print('WebSocket connection closed');
    // Implement reconnection logic here
  }

  void dispose() {
    _channel.sink.close();
  }
}
```

To use this WebSocket service in your Flutter app:

```dart
class MyApp extends StatefulWidget {
  @override
  _MyAppState createState() => _MyAppState();
}

class _MyAppState extends State<MyApp> {
  final WebSocketService _webSocketService = WebSocketService();

  @override
  void initState() {
    super.initState();
    _webSocketService.connect();
  }

  @override
  void dispose() {
    _webSocketService.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Your app's widget tree
  }
}
```

Remember to handle WebSocket connection errors, implement reconnection logic, and consider using a state management solution like Provider, Riverpod, or BLoC to update your UI based on these messages. Also, ensure that you're handling potential security considerations in your actual implementation.
