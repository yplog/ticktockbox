# TickTockBox API Documentation

This document provides detailed information about the TickTockBox REST API endpoints.

## Base URL

```
{url}:{port}
```

## Authentication

Currently, the API does not require authentication. This may change in future versions.

## Content Type

All API requests that include a body should use `application/json` content type.

## Response Format

All API responses follow a consistent format:

### Success Response
```json
{
  "success": true,
  "data": <response_data>,
  "message": "<optional_message>"
}
```

### Error Response
```json
{
  "success": false,
  "message": "<error_message>"
}
```

## Endpoints

### 1. Create Message

Creates a new scheduled message.

**Endpoint:** `POST /api/messages`

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "message": "string",
  "expire_at": "string (ISO 8601 timestamp)"
}
```

**Parameters:**
- `message` (required): The message content to be scheduled
- `expire_at` (required): ISO 8601 timestamp when the message should expire (must be in the future)

**Example Request:**
```bash
curl -X POST {url}:{port}/api/messages \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Reminder: Meeting at 3 PM",
    "expire_at": "2025-05-26T15:00:00Z"
  }'
```

**Success Response (201 Created):**
```json
{
  "success": true,
  "message": "Message created successfully"
}
```

**Error Responses:**

*400 Bad Request - Invalid JSON:*
```json
{
  "success": false,
  "message": "Invalid JSON format"
}
```

*400 Bad Request - Missing fields:*
```json
{
  "success": false,
  "message": "Message and expire_at are required"
}
```

*400 Bad Request - Invalid timestamp:*
```json
{
  "success": false,
  "message": "Invalid expire_at format. Use ISO 8601 format (e.g., 2024-01-01T12:00:00Z)"
}
```

*400 Bad Request - Past timestamp:*
```json
{
  "success": false,
  "message": "ExpireAt must be in the future"
}
```

*500 Internal Server Error:*
```json
{
  "success": false,
  "message": "Failed to create message"
}
```

### 2. Get All Active Messages

Retrieves all messages that have not yet been processed (not expired or not yet delivered).

**Endpoint:** `GET /api/messages`

**Parameters:** None

**Example Request:**
```bash
curl {url}:{port}/api/messages
```

**Success Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": 1748284721605344000,
      "message": "Reminder: Meeting at 3 PM",
      "expire_at": "2025-05-26T15:00:00Z",
      "created": "2025-05-26T14:30:24.249571Z"
    },
    {
      "id": 1748284721605344001,
      "message": "System maintenance in 1 hour",
      "expire_at": "2025-05-26T16:00:00Z",
      "created": "2025-05-26T14:45:12.123456Z"
    }
  ]
}
```

**Success Response (200 OK) - No messages:**
```json
{
  "success": true,
  "data": null
}
```

**Response Fields:**
- `id`: Unique message identifier (int64)
- `message`: The message content
- `expire_at`: ISO 8601 timestamp when the message will expire
- `created`: ISO 8601 timestamp when the message was created

**Error Response (500 Internal Server Error):**
```json
{
  "success": false,
  "message": "Failed to retrieve messages"
}
```

### 3. Health Check

Check if the API is running and healthy.

**Endpoint:** `GET /health`

**Example Request:**
```bash
curl {url}:{port}/health
```

**Success Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2025-05-26T14:30:00Z"
}
```

## WebSocket API

### Connection

Connect to the WebSocket endpoint to receive real-time notifications when messages expire.

**Endpoint:** `ws://{url}:{port}/ws`

**Example JavaScript:**
```javascript
const ws = new WebSocket('ws://{url}:{port}/ws');

ws.onopen = function(event) {
    console.log('Connected to WebSocket');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received expired message:', data);
};

ws.onclose = function(event) {
    console.log('WebSocket connection closed');
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};
```

### Message Format

When a message expires, all connected WebSocket clients receive a notification:

```json
{
  "type": "expired_messages",
  "data": [
    {
      "id": 1748284721605344000,
      "message": "Reminder: Meeting at 3 PM",
      "expire_at": "2025-05-26T15:00:00Z",
      "created": "2025-05-26T14:30:24.249571Z"
    }
  ],
  "timestamp": "2025-05-26T15:00:01Z"
}
```

## Rate Limiting

Currently, there are no rate limits implemented. This may change in future versions.

## Error Handling

### HTTP Status Codes

- `200 OK`: Request successful
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `404 Not Found`: Endpoint not found
- `500 Internal Server Error`: Server error

### Common Error Scenarios

1. **Invalid JSON**: Malformed JSON in request body
2. **Missing Required Fields**: Required fields not provided
3. **Invalid Timestamp Format**: Timestamp not in ISO 8601 format
4. **Past Timestamp**: Expiration time is in the past
5. **Database Connection Error**: Unable to connect to QuestDB
6. **Internal Server Error**: Unexpected server error

## Examples

### Complete Workflow Example

```bash
# 1. Create a message that expires in 5 minutes
curl -X POST {url}:{port}/api/messages \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Test notification",
    "expire_at": "2025-05-26T15:05:00Z"
  }'

# 2. Check all active messages
curl {url}:{port}/api/messages

# 3. Wait for the message to expire (5 minutes)
# The message will be automatically delivered via WebSocket and RabbitMQ

# 4. Check active messages again (should be empty or not include the expired message)
curl {url}:{port}/api/messages
```

### Batch Message Creation

```bash
# Create multiple messages
for i in {1..5}; do
  expire_time=$(date -u -d "+${i} minutes" +"%Y-%m-%dT%H:%M:%SZ")
  curl -X POST {url}:{port}/api/messages \
    -H "Content-Type: application/json" \
    -d "{
      \"message\": \"Message ${i}\",
      \"expire_at\": \"${expire_time}\"
    }"
done
```
