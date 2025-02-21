# Balloon Popping Game

An interactive game designed to generate streaming data for real-time analytics demonstrations. The game uses WebSocket for real-time interactions and streams events to Kafka, making it perfect for showcasing streaming databases and real-time analytics capabilities.

## Features & Use Cases

-   Real-time player performance analytics
-   Time-series analysis of game events
-   Pattern detection in player behavior
-   Color popularity and bonus effectiveness tracking
-   Visual effects for bonus balloons
-   Real-time score updates with animations

---

## Getting Started

### Prerequisites

-   Docker and Docker Compose (recommended)
-   Web browser with HTML5 support

### Quick Start with Docker

1. Clone and start the application:

```bash
git clone <repository-url>
cd balloon-game
docker compose up -d
```

2. Start a game session:

```bash
curl -X POST http://localhost:8000/game/start
```

3. Open http://localhost:8000 in your browser

    - Enter player name
    - Select character
    - Start playing

4. Monitor events:

```bash
docker compose exec kafka \
  bin/kafka-console-consumer.sh \
    --bootstrap-server localhost:9092 \
    --topic game_scores \
    --from-beginning
```

### Manual Setup (Without Docker)

1. Create environment file:

```bash
# .env
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
KAFKA_TOPIC=game-scores
APP_LOG_LEVEL=DEBUG
```

2. Install dependencies:

```bash
uv venv
source .venv/bin/activate  # Linux/Mac
uv pip install -r requirements.txt
```

3. Run the application:

```bash
uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

---

## Game Details

### Character Configuration

Each character has favorite colors that give bonus points:

```python
{
    "Jerry": ["brown", "yellow"],
    "Mickey": ["red", "black"],
    "Sonic": ["blue", "gold"],
    # ... more characters
}
```

### Scoring System

-   Base points vary by balloon color
-   Bonus points (2x) for favorite colors
-   Visual sparkle effect for bonus balloons
-   Real-time score animations

### Game Controls

```bash
# Start game
curl -X POST http://localhost:8000/game/start

# Check status
curl http://localhost:8000/game/status

# Stop game
curl -X POST http://localhost:8000/game/stop
```

---

## Technical Documentation

### Architecture

#### Backend

-   [FastAPI](https://fastapi.tiangolo.com/) - Modern web framework
-   [Uvicorn](https://www.uvicorn.org/) - ASGI server
-   [WebSockets](https://websockets.readthedocs.io/) - Real-time communication
-   [Pydantic](https://docs.pydantic.dev/) - Data validation

#### Event Streaming

-   [Apache Kafka](https://kafka.apache.org/) - Event streaming platform
-   [aiokafka](https://aiokafka.readthedocs.io/) - Async Kafka client

#### Frontend

-   [HTML5 Canvas](https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API) - Game rendering
-   [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket) - Real-time updates

### Data Streams

#### Score Events

```json
{
    "player": "PlayerName",
    "balloon_color": "red",
    "score": 100,
    "favorite_color_bonus": true,
    "event_ts": "2025-02-04T12:25:35.202Z"
}
```

#### Session Events

```json
{
    "event_type": "session",
    "action": "start|stop",
    "player_count": 5,
    "session_id": "uuid",
    "event_ts": "2025-02-04T12:25:35.202Z"
}
```

### Stream Characteristics

-   Topic: `game_scores` (configurable)
-   Timestamp precision: Milliseconds
-   Event ordering: Preserved within sessions
-   Time zone: UTC
-   Data rate: Variable based on active players

---

## Development Guide

### Project Structure

```
balloon-game/
├── main.py              # FastAPI application
├── models.py            # Data models
├── kafka_producer.py    # Kafka integration
├── static/             # Frontend assets
├── docker-compose.yml  # Docker configuration
└── requirements.txt    # Dependencies
```

### Development Tools

-   [uv](https://github.com/astral-sh/uv) - Package management
-   [ruff](https://github.com/astral-sh/ruff) - Linting and formatting
-   [mockafka-py](https://github.com/notnot/mockafka) - Testing

### Docker Development

```bash
# Rebuild after changes
docker compose build game
docker compose up -d game

# View logs
docker compose logs -f

# Clean up
docker compose down -v
```

### Environment Variables

```yaml
KAFKA_BOOTSTRAP_SERVERS: kafka:9092
KAFKA_TOPIC: game_scores
APP_LOG_LEVEL: DEBUG
BONUS_PROBABILITY: 0.15
```

### Troubleshooting

#### WebSocket Issues

-   Check FastAPI server status
-   Verify WebSocket URL
-   Check browser console
-   Review DEBUG logs

#### Kafka Issues

-   Verify broker is running
-   Check connection strings
-   Confirm topic exists
-   Review broker logs

#### Game Issues

-   Verify game session is active
-   Check API responses
-   Monitor WebSocket state
-   Review browser console

---

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## Related Documentation

-   [FastAPI WebSocket Guide](https://fastapi.tiangolo.com/advanced/websockets/)
-   [Kafka Documentation](https://kafka.apache.org/documentation/)
-   [HTML5 Canvas Tutorial](https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial)
-   [Pydantic v2 Documentation](https://docs.pydantic.dev/latest/)
