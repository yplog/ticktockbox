# TickTockBox

<p align="center">
  <img src="https://ucb74ee65c62f925367ee1d00913.previews.dropboxusercontent.com/p/thumb/ACqbJcsoxLJDLIa442I7_pbtaVNKAVYQkIwS4qdUa0NdP2CwkC7ruM0HilUUokhvvuPM_oyt3s5CmrVlo2vziLvocxs7gQXSeQ17E8zIjsx_pmdUYf_1gI1JTZWrsMk3lkogo-3QJvkoJ4FsACMKbIx80qqZIlIBdzjYe3bS1y8bJj1uePss2kmXHmJ_uKYaf0jZu0B_rWPV0f6NudEFmnvISegIwSVysf7VNTdcCqERT9eXzj7fR_m6PAIjMHO3HEGrPVpUYkrDMtQfhC1gEgyWsfnBZUz4UC9JlXgzn_tijOJVx7-KHA8D5lRBjomnAYbbqVa4W7WqrNFyasHm-jg550uMkIeRNj04JJV6pwHyN4J4nhd1RA7H-_7yavm46_E/p.png?is_prewarmed=true" width="256" height="256" alt="TickTockBox">
</p>

A high-performance reminder and job scheduling system built with Go. TickTockBox uses a custom timing wheel algorithm to efficiently manage thousands of scheduled tasks, delivering notifications through RabbitMQ when reminders are due.

## Features

- **Timing Wheel Scheduler**: Custom implementation for high-performance task scheduling
- **Web Admin Interface**: User-friendly web UI for creating and managing reminders
- **DateTime Picker**: Modern date/time selection with timezone support
- **Multi-Timezone Support**: 20+ timezone options with automatic UTC conversion
- **SQLite Database**: Lightweight, embedded database for persistence
- **RabbitMQ Integration**: Reliable message delivery when reminders are due

## Architecture

TickTockBox implements a **hierarchical timing wheel** algorithm similar to those used in production systems like Redis and Kafka:

- **Circular Array**: Buckets organized in a circular array structure
- **Hash-based Distribution**: Timers distributed across buckets based on delay
- **Round-based Scheduling**: Long-term timers handled with round counters
- **Batch Processing**: Efficient processing of multiple timers per tick

## Quick Start

### Prerequisites

- Go 1.25+
- RabbitMQ (optional, for message delivery)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yplog/ticktockbox.git
cd ticktockbox
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run ./cmd/server/main.go
```

4. Open your browser and visit `http://localhost:8080`

### Using Docker Compose

1. Start RabbitMQ (optional):
```bash
docker run -d --name rabbitmq \
  -p 5672:5672 -p 15672:15672 \
  rabbitmq:3-management
```

2. Set environment variables:
```bash
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export RABBITMQ_QUEUE="reminders.due"
```

## Usage

### Web Interface

1. **Create Reminders**: Navigate to `/new` to create a new reminder
2. **Select Timezone**: Choose from 20+ predefined timezones
3. **Pick Date/Time**: Use the modern datetime picker for precise scheduling
4. **Set Reminder Time**: Configure how many minutes before the event to be reminded
5. **View Upcoming**: See all pending reminders on the main dashboard

### API Examples

Create a reminder programmatically:
```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "title=Team Meeting" \
  -d "tz=Europe/Istanbul" \
  -d "run_at=2025-09-06T14:30:00" \
  -d "remind_before_minutes=15"
```

## Configuration

Configure using environment variables:

```bash
# Server configuration
export ADDR=":8080"                                    # Server address
export SQLITE_PATH="app.db"                           # SQLite database path

# RabbitMQ configuration (optional)
export RABBITMQ_URL="amqp://guest:guest@localhost:5672/"
export RABBITMQ_QUEUE="reminders.due"

# Timing wheel configuration (optional)
export WHEEL_TICK="1s"                                # Tick duration
export WHEEL_SLOTS="512"                              # Number of slots
```

## Timing Wheel Algorithm

The core of TickTockBox is a custom timing wheel implementation:

```go
// Create a timing wheel with 1-second ticks and 512 slots
wheel := twheel.New(1*time.Second, 512)
wheel.Start()

// Schedule a task
id := wheel.AfterFunc(5*time.Minute, func() {
    fmt.Println("Reminder fired!")
})

// Cancel if needed
wheel.Cancel(id)
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- [Create an issue](https://github.com/yplog/ticktockbox/issues) on GitHub
- Check the code documentation
- Review application logs for debugging
