# TickTockBox

<p align="center">
  <img src="https://ucb74ee65c62f925367ee1d00913.previews.dropboxusercontent.com/p/thumb/ACqbJcsoxLJDLIa442I7_pbtaVNKAVYQkIwS4qdUa0NdP2CwkC7ruM0HilUUokhvvuPM_oyt3s5CmrVlo2vziLvocxs7gQXSeQ17E8zIjsx_pmdUYf_1gI1JTZWrsMk3lkogo-3QJvkoJ4FsACMKbIx80qqZIlIBdzjYe3bS1y8bJj1uePss2kmXHmJ_uKYaf0jZu0B_rWPV0f6NudEFmnvISegIwSVysf7VNTdcCqERT9eXzj7fR_m6PAIjMHO3HEGrPVpUYkrDMtQfhC1gEgyWsfnBZUz4UC9JlXgzn_tijOJVx7-KHA8D5lRBjomnAYbbqVa4W7WqrNFyasHm-jg550uMkIeRNj04JJV6pwHyN4J4nhd1RA7H-_7yavm46_E/p.png?is_prewarmed=true" width="256" height="256" alt="TickTockBox">
</p>

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
# Start dependencies: QuestDB + RabbitMQ
make deps-up
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

## Configuration

TickTockBox uses environment variables for configuration. Copy `env.example` to `.env` and modify as needed:

### Event Types

- **`created`**: Message was created
- **`processed`**: Message was processed and expired

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
