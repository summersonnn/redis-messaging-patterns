# 🚀 Redis Messaging Patterns Demo

![Docker](https://img.shields.io/badge/Docker-20.10+-blue?logo=docker)
![Python](https://img.shields.io/badge/Python-3.9+-green?logo=python)
![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![Redis](https://img.shields.io/badge/Redis-7.0+-red?logo=redis)
![License](https://img.shields.io/badge/License-MIT-yellow)

> **A comprehensive demonstration of Redis messaging patterns using Python and Go in a containerized environment**

This project showcases two fundamental Redis messaging patterns - **Pub/Sub** and **Queue-based messaging** - through a practical implementation featuring a Python publisher service and a Go subscriber service, all orchestrated with Docker Compose.

## 📋 Table of Contents

- [Overview](#-overview)
- [Architecture](#️-architecture)  
- [Quick Start](#-quick-start)
- [Messaging Patterns](#-messaging-patterns)
- [Configuration](#-configuration)
- [Project Structure](#-project-structure)
- [Output Examples](#-output-examples)
- [Troubleshooting](#-troubleshooting)

## 🎯 Overview

### What This Project Demonstrates

- **Dual Messaging Patterns**: Compare Pub/Sub vs Queue messaging in action
- **Polyglot Implementation**: Python publisher + Go subscriber showcasing cross-language Redis integration
- **Production-Ready Features**: Connection retry logic, graceful shutdowns, structured logging
- **Container Orchestration**: Complete Docker Compose setup with health checks and networking

### Use Cases

- **Learning**: Understand Redis messaging patterns hands-on
- **Prototyping**: Template for building distributed messaging systems  
- **Comparison**: See the differences between broadcast and queue-based messaging
- **Integration**: Example of Python-Go service communication via Redis

## 🏗️ Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│                     │    │                     │    │                     │
│   Python Publisher  │    │       Redis         │    │   Go Subscriber     │
│                     │───▶│   Message Broker    │───▶│                     │
│   • Generates msgs  │    │   • Pub/Sub         │    │   • Consumes msgs   │
│   • UUID + metadata │    │   • Queue/List      │    │   • Concurrent      │
│   • Configurable    │    │   • Persistence     │    │   • Thread-safe     │
│                     │    │                     │    │                     │
└─────────────────────┘    └─────────────────────┘    └─────────────────────┘
       (Container)               (Container)               (Container)
```

### Data Flow

1. **Python Publisher** generates structured messages with UUIDs and timestamps
2. **Redis** receives messages via two patterns:
   - **Channel**: `messages` (Pub/Sub pattern)
   - **List**: `message_queue` (Queue pattern)
3. **Go Subscriber** consumes from both sources concurrently using goroutines

## 🚀 Quick Start

### Prerequisites

- ✅ **Docker** and **Docker Compose** installed
- ✅ Port **6379** available (Redis default)

### 1. Clone and Run

```bash
# Clone the repository
git clone https://github.com/summersonnn/redis-messaging-patterns.git
cd redis-go

# Start all services (both patterns)
docker-compose up --build
```

### 2. Choose Messaging Pattern

```bash
# 📡 Pub/Sub only (real-time broadcast)
PATTERN=pubsub docker-compose up --build

# 📥 Queue only (persistent FIFO)  
PATTERN=queue docker-compose up --build

# 🔄 Both patterns (default)
PATTERN=both docker-compose up --build
```

> **⚠️ Important:** When changing the `PATTERN` environment variable, make sure to run `docker-compose down` before running `docker-compose up` again to ensure the configuration change takes effect properly.

### 3. Watch the Magic

You'll see real-time output showing messages flowing between services:

```
python-publisher_1  | [2025-07-29T10:30:15.123] Published to channel 'messages' (1 subscribers): Hello from Python #1
go-subscriber_1     | [2025-07-29 10:30:15] [python-publisher] Hello from Python #1 (ID: abc-123) [Received #1 from pub-sub]
```

### 4. Clean Up

```bash
# Stop all services and remove containers
docker-compose down

# Remove volumes (clears Redis data)
docker-compose down -v
```

## 🔄 Messaging Patterns

### 📡 Pub/Sub Pattern (`PUBLISH` / `SUBSCRIBE`)

**How it works:**
```
Publisher ──PUBLISH──▶ Redis Channel ──SUBSCRIBE──▶ Subscriber(s)
```

**Characteristics:**
- ⚡ **Real-time**: Messages delivered instantly to all active subscribers
- 🗂️ **No persistence**: Messages are lost if no subscribers are listening
- 📢 **Broadcast**: Every subscriber receives every message
- 🎯 **Use case**: Live notifications, real-time updates, event broadcasting

**Redis Commands:**
- Publisher: `PUBLISH messages "{json_message}"`
- Subscriber: `SUBSCRIBE messages`

### 📥 Queue Pattern (`RPUSH` / `BLPOP`)

**How it works:**
```
Publisher ──RPUSH──▶ Redis List ──BLPOP──▶ Subscriber
```

**Characteristics:**
- 💾 **Persistent**: Messages saved until consumed
- 📋 **FIFO**: First In, First Out processing order
- ⚖️ **Load balancing**: Multiple subscribers can share work
- 🎯 **Use case**: Job queues, task processing, background jobs

**Redis Commands:**
- Publisher: `RPUSH message_queue "{json_message}"`
- Subscriber: `BLPOP message_queue 1` (1-second timeout)

### 🆚 Pattern Comparison

| Feature | Pub/Sub | Queue |
|---------|---------|-------|
| **Persistence** | ❌ No | ✅ Yes |
| **Message Loss** | Possible | Guaranteed delivery |
| **Scalability** | Fan-out | Load distribution |
| **Latency** | Ultra-low | Low |
| **Durability** | None | Redis durability |

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `redis` | Redis server hostname |
| `REDIS_PORT` | `6379` | Redis server port |
| `PATTERN` | `both` | Messaging pattern (`pubsub`, `queue`, `both`) |
| `PUBLISH_INTERVAL` | `2.0` | Seconds between published messages |

### Custom Configuration Examples

```bash
# High-frequency publishing (0.5s intervals)
PUBLISH_INTERVAL=0.5 docker-compose up

# External Redis instance
REDIS_HOST=my-redis.example.com REDIS_PORT=6380 docker-compose up

# Queue-only pattern with custom interval
PATTERN=queue PUBLISH_INTERVAL=1.0 docker-compose up
```

## 📁 Project Structure

```
redis-go/
├── 🐳 docker-compose.yml           # Service orchestration
├── 📖 README.md                    # This file
├── 📝 CLAUDE.md                    # Project guidance
├── 🐍 python-publisher/            # Python message publisher
│   ├── 🐳 Dockerfile              # Python service container
│   ├── 🐍 main.py                 # Publisher implementation
│   └── 📦 requirements.txt        # Python dependencies (redis==5.0.1)
└── 🔧 go-subscriber/               # Go message subscriber
    ├── 🐳 Dockerfile              # Go service container  
    ├── 📦 go.mod                  # Go module definition
    └── 🔧 main.go                 # Subscriber implementation
```

### Key Files

- **`docker-compose.yml`**: Defines Redis, Python publisher, and Go subscriber services
- **`python-publisher/main.py`**: Message generator with configurable patterns
- **`go-subscriber/main.go`**: Concurrent message consumer
- **Dockerfiles**: Multi-stage builds for optimized containers

## 📊 Output Examples

### Console Output

When running with `PATTERN=both`, you'll see output like:

```bash
# Python Publisher
python-publisher_1  | [2025-07-29T10:30:15.123] Starting Python Publisher...
python-publisher_1  | [2025-07-29T10:30:15.124] Connected to Redis at redis:6379
python-publisher_1  | [2025-07-29T10:30:15.125] Published to channel 'messages' (1 subscribers): Hello from Python #1
python-publisher_1  | [2025-07-29T10:30:15.126] Pushed to queue 'message_queue' (length: 1): Hello from Python #1

# Go Subscriber  
go-subscriber_1     | [2025-07-29T10:30:15.127] Starting Go Subscriber...
go-subscriber_1     | [2025-07-29T10:30:15.128] Connected to Redis at redis:6379
go-subscriber_1     | [2025-07-29 10:30:15] [python-publisher] Hello from Python #1 (ID: abc-123) [Received #1 from pub-sub]
go-subscriber_1     | [2025-07-29 10:30:15] [python-publisher] Hello from Python #1 (ID: abc-123) [Received #2 from queue]
```

### Message Structure

Each message follows this JSON schema:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-07-29T10:30:15.123456Z", 
  "message": "Hello from Python #42",
  "sender": "python-publisher",
  "count": 42
}
```

## 🔧 Troubleshooting

### Common Issues

#### 🚫 Port 6379 Already in Use
```bash
# Check what's using the port
sudo lsof -i :6379

# Stop local Redis if running
sudo systemctl stop redis-server

# Or use different port
REDIS_PORT=6380 docker-compose up
```

#### 🔄 Services Not Connecting  
```bash
# Check service health
docker-compose ps

# View service logs
docker-compose logs redis
docker-compose logs python-publisher
docker-compose logs go-subscriber

# Restart with fresh containers
docker-compose down && docker-compose up --build
```

#### 📊 No Messages Appearing
```bash
# Verify pattern configuration
echo $PATTERN  # Should be 'pubsub', 'queue', or 'both'

# Check Redis connectivity
docker-compose exec redis redis-cli ping
# Should return: PONG

# Monitor Redis activity
docker-compose exec redis redis-cli monitor
```

#### 🐳 Container Build Issues
```bash
# Clean Docker cache
docker system prune -a

# Rebuild without cache
docker-compose build --no-cache

# Check Docker resources
docker system df
```

### Advanced Debugging

#### Access Redis CLI
```bash
# Connect to Redis container
docker-compose exec redis redis-cli

# Monitor pub/sub activity
MONITOR

# Check queue length
LLEN message_queue

# View recent messages
LRANGE message_queue 0 -1
```

#### Performance Monitoring
```bash
# View container stats
docker stats

# Monitor Redis performance
docker-compose exec redis redis-cli --stat

# Check memory usage
docker-compose exec redis redis-cli info memory
```

## 🎓 Learning Outcomes

After running this demo, you'll understand:

- ✅ How Redis Pub/Sub differs from queue-based messaging
- ✅ When to use each pattern in real applications  
- ✅ Cross-language service communication via Redis
- ✅ Container orchestration with Docker Compose
- ✅ Production patterns: retry logic, graceful shutdowns, health checks
- ✅ Message serialization and structured logging
