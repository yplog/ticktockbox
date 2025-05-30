# TickTockBox

A high-performance time-based message scheduling and notification system built with Go, QuestDB, and RabbitMQ. TickTockBox schedules messages with expiration times and automatically processes them when they expire, delivering notifications through RabbitMQ message queues and optional real-time WebSocket updates.

## Features

- **Time-based Message Scheduling**: Schedule messages with specific expiration times
- **Message Queue Integration**: Automatic publishing to RabbitMQ when messages expire
- **Real-time Notifications**: WebSocket-based real-time updates when messages expire
- **High-Performance Storage**: QuestDB time-series database with automatic partitioning

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24.3+ (for development)

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/yplog/ticktockbox.git
cd ticktockbox
```

2. Start the services:
```bash
docker compose up -d
```

This will start:
- QuestDB on ports 9000 (HTTP), 8812 (PostgreSQL), 9009 (InfluxDB)
- RabbitMQ on ports 5672 (AMQP), 15672 (Management UI)

3. Build and run TickTockBox:
```bash
make build
make run
```

4. Test the API:
```bash
# Create a message
curl -X POST http://localhost:3000/api/messages \
  -H "Content-Type: application/json" \
  -d '{"message": "Test message", "expire_at": "2025-01-15T10:30:00Z"}'

# Get all messages
curl http://localhost:3000/api/messages
```

### Manual Setup

1. Start QuestDB and RabbitMQ:
```bash
# QuestDB
docker run -p 9000:9000 -p 8812:8812 -p 9009:9009 questdb/questdb:latest

# RabbitMQ
docker run -p 5672:5672 -p 15672:15672 rabbitmq:3-management
```

2. Configure environment variables:
```bash
cp env.example .env
# Edit .env file as needed
```

3. Build and run:
```bash
go mod download
go build -o bin/ticktockbox cmd/server/main.go
./bin/ticktockbox
```

## Configuration

TickTockBox uses environment variables for configuration. Copy `env.example` to `.env` and modify as needed:

```env
# Server Configuration
PORT=3000

# QuestDB Configuration
QUESTDB_URL=localhost:8812
QUESTDB_USER=admin
QUESTDB_PASSWORD=quest
QUESTDB_NAME=qdb
QUESTDB_SSL_MODE=disable
QUESTDB_HTTP_URL=http://localhost:9000

# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

### Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | HTTP server port |
| `QUESTDB_URL` | `localhost:8812` | QuestDB PostgreSQL wire protocol URL |
| `QUESTDB_USER` | `admin` | QuestDB username |
| `QUESTDB_PASSWORD` | `quest` | QuestDB password |
| `QUESTDB_NAME` | `qdb` | QuestDB database name |
| `QUESTDB_SSL_MODE` | `disable` | SSL mode: disable, require, verify-ca, verify-full |
| `QUESTDB_HTTP_URL` | `http://localhost:9000` | QuestDB HTTP API URL |
| `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | RabbitMQ AMQP connection URL |

## API Reference

### Create Message

Create a new scheduled message.

```http
POST /api/messages
Content-Type: application/json

{
  "message": "Your message content",
  "expire_at": "2024-01-15T10:30:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Message created successfully"
}
```

### Get Messages

Retrieve all messages with their current status.

```http
GET /api/messages
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "message_id": 1705312200000000000,
      "message": "Your message content",
      "expire_at": "2024-01-15T10:30:00Z",
      "event_type": "created",
      "timestamp": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### WebSocket Connection

Connect to real-time updates for expired messages:

```javascript
const ws = new WebSocket('ws://localhost:3000/ws');

ws.onopen = function() {
    console.log('Connected to WebSocket');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    if (data.type === 'expired_messages') {
        console.log('Expired messages received:', data.messages);
    }
};

ws.onclose = function() {
    console.log('WebSocket connection closed');
};
```

## Message Lifecycle

1. **Creation**: Messages are created with a future expiration time
2. **Storage**: Messages are stored in QuestDB with "created" event type
3. **Monitoring**: Scheduler checks for expired messages every 10 seconds
4. **Processing**: When messages expire:
   - Published to RabbitMQ exchange "expired_messages"
   - Broadcasted via WebSocket to connected clients
   - Event type updated to "processed" in database
5. **Cleanup**: Old partitions are automatically cleaned up daily

## Event-Driven Architecture

TickTockBox uses QuestDB's append-newest pattern for event sourcing:

### Database Schema

```sql
CREATE TABLE messages (
    ts TIMESTAMP,           -- Event timestamp
    message_id LONG,        -- Unique message identifier
    message STRING,         -- Message content
    expire_at TIMESTAMP,    -- Expiration time
    event_type SYMBOL      -- Event type (created/processed)
) timestamp(ts) PARTITION BY DAY;
```

### Event Types

- **`created`**: Message was created
- **`processed`**: Message was processed and expired

### Latest State Query

```sql
WITH latest_messages AS (
    SELECT max(ts) AS ts, message_id FROM messages GROUP BY message_id
)
SELECT m.* FROM latest_messages lm
INNER JOIN messages m ON lm.ts = m.ts AND lm.message_id = m.message_id
```

## RabbitMQ Integration

TickTockBox publishes expired messages to a RabbitMQ fanout exchange:

- **Exchange**: `expired_messages`
- **Type**: `fanout`
- **Routing**: Broadcasts to all bound queues
- **Message Format**: JSON with complete message data

### Consuming Messages

To consume expired messages from RabbitMQ:

```go
// Declare a queue and bind to the exchange
queue, err := channel.QueueDeclare("", false, false, true, false, nil)
err = channel.QueueBind(queue.Name, "", "expired_messages", false, nil)

// Consume messages
msgs, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
```

## Development

### Project Structure

```
ticktockbox/
├── cmd/
│   └── server/          # Application entry point
├── internal/
│   ├── api/            # HTTP API handlers
│   ├── config/         # Configuration management
│   ├── database/       # QuestDB integration
│   ├── rabbitmq/       # RabbitMQ integration
│   ├── scheduler/      # Background scheduler
│   └── websocket/      # WebSocket hub
├── docker-compose.yml  # Development services
├── Dockerfile         # Application container
└── Makefile          # Build automation
```

### Available Make Commands

```bash
make build          # Build the application
make run            # Run the application
make test           # Run tests
make clean          # Clean build artifacts
make docker-build   # Build Docker image
make docker-run     # Run with Docker
make dev            # Start development environment
```

### Running Tests

```bash
make test
```

### Development with Air (Hot Reload)

For development with automatic reloading:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## Monitoring

### Health Checks

TickTockBox includes built-in health checks:

```bash
# Docker health check
docker ps  # Check container health status

# Manual health check (if implemented)
curl http://localhost:3000/health
```

### QuestDB Console

Access QuestDB web console at `http://localhost:9000` to:
- Monitor message storage
- Run custom queries
- View partition information
- Check system performance

### RabbitMQ Management

Access RabbitMQ management UI at `http://localhost:15672`:
- Username: `guest`
- Password: `guest`

Monitor:
- Message throughput
- Queue status
- Exchange bindings
- Connection status

## Production Deployment

### Docker Deployment

1. Build the production image:
```bash
make docker-build
```

2. Use the full Docker Compose setup:
```bash
docker-compose -f docker-compose.full.yml up -d
```

### Environment Considerations

- **Security**: Change default RabbitMQ credentials
- **SSL/TLS**: Configure proper SSL settings for production
- **Persistence**: Ensure data volumes are properly configured
- **Networking**: Configure proper network security
- **Monitoring**: Set up logging and monitoring solutions
- **Backup**: Implement backup strategies for QuestDB data

### Scaling

- **Horizontal Scaling**: Multiple TickTockBox instances can run simultaneously
- **Database**: QuestDB handles high-throughput time-series data efficiently
- **Message Queue**: RabbitMQ provides reliable message delivery and can be clustered
- **Load Balancing**: Use a load balancer for multiple application instances

## Troubleshooting

### Common Issues

1. **Connection Refused Errors**
   - Ensure QuestDB and RabbitMQ are running
   - Check port availability
   - Verify connection URLs in configuration

2. **WebSocket Connection Issues**
   - Check CORS configuration
   - Verify WebSocket endpoint accessibility
   - Check browser console for errors

3. **Message Not Processing**
   - Verify scheduler is running
   - Check QuestDB connectivity
   - Review application logs

4. **SSL/TLS Issues**
   - Verify SSL mode configuration
   - Check certificate validity
   - Ensure proper SSL setup

### Logs

Application logs provide detailed information about:
- Scheduler operations
- Database operations
- WebSocket connections
- RabbitMQ publishing
- Error conditions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the logs for error details
