# Delivery System Backend

A production-ready backend foundation for a delivery system built with Go, featuring robust authentication, session management, WebSockets, and message queue infrastructure.

## 🚀 Features

- **Authentication System**: Complete user registration, login, and session management
- **Session Management**: Redis-backed sessions with secure HTTP-only cookies
- **WebSocket Support**: Real-time communication with cookie-based authentication
- **Message Queue**: Event-driven architecture using Redis Streams
- **Security**: Rate limiting, CORS, bcrypt password hashing
- **Clean Architecture**: Well-structured, modular, and maintainable codebase

## 📋 Tech Stack

- **Language**: Go 1.22+
- **Web Framework**: Chi Router
- **Database**: PostgreSQL 15+
- **Cache/Session Store**: Redis 7+
- **Message Queue**: Redis Streams
- **WebSockets**: Gorilla WebSocket
- **Security**: bcrypt, secure cookies, rate limiting

## 📁 Project Structure

```
kovadelivery.com/
├── cmd/api/                    # Application entry point
│   └── main.go
├── internal/                   # Private application code
│   ├── config/                 # Configuration management
│   ├── db/                     # Database connection
│   ├── cache/                  # Redis client
│   ├── auth/                   # Authentication logic
│   │   ├── service.go
│   │   ├── session.go
│   │   └── password.go
│   ├── http/                   # HTTP layer
│   │   ├── router.go
│   │   ├── middleware/         # Auth, CORS, rate limiting
│   │   └── handlers/           # Request handlers
│   ├── models/                 # Data models
│   ├── mq/                     # Message queue
│   │   ├── producer.go
│   │   ├── consumer.go
│   │   └── events.go
│   └── ws/                     # WebSocket
│       ├── hub.go
│       ├── client.go
│       └── handler.go
├── pkg/utils/                  # Shared utilities
├── migrations/                 # Database migrations
└── .env.example               # Environment template
```

## 🛠️ Setup

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 15+
- Redis 7+
- Make (optional)

### Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd kovadelivery.com
```

2. **Install dependencies**
```bash
go mod download
```

3. **Setup environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Start dependencies with Docker**
```bash
docker-compose up -d
```

Or install PostgreSQL and Redis manually.

5. **Run database migrations**
```bash
# Using psql
psql postgresql://postgres:postgres@localhost:5432/kova_delivery -f migrations/001_create_users_table.up.sql
psql postgresql://postgres:postgres@localhost:5432/kova_delivery -f migrations/002_create_sessions_table.up.sql

# Or using make
make migrate-up DB_URL=postgresql://postgres:postgres@localhost:5432/kova_delivery
```

6. **Run the application**
```bash
# Using go
go run cmd/api/main.go

# Or using make
make run

# Or build and run
make build
./bin/api
```

The server will start on `http://localhost:8080`

## 🔑 Environment Configuration

Key environment variables (see `.env.example` for full list):

```bash
# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=kova_delivery

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Session (IMPORTANT: Change in production!)
SESSION_SECRET=change-this-to-a-random-64-character-string-in-production
SESSION_DURATION=24h

# Security
BCRYPT_COST=12
COOKIE_SECURE=false  # Set to true in production with HTTPS
```

## 📡 API Endpoints

### Authentication

**Register User**
```bash
POST /api/auth/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "securepassword123",
  "role": "customer"  # Optional: customer, driver, admin
}
```

**Login**
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword123"
}

# Returns session cookie in response
```

**Logout**
```bash
POST /api/auth/logout
Cookie: session_id=<session-id>
```

### Protected Routes

**Get Current User**
```bash
GET /api/users/me
Cookie: session_id=<session-id>
```

### WebSocket

**Connect to WebSocket**
```bash
GET /api/ws
Cookie: session_id=<session-id>
Upgrade: websocket
```

### Health Check

```bash
GET /health
```

## 🧪 Testing

### Manual Testing with cURL

**Register a user:**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "phone": "+1234567890",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Get current user:**
```bash
curl -X GET http://localhost:8080/api/users/me \
  -b cookies.txt
```

**Connect to WebSocket** (using a WebSocket client):
```javascript
const ws = new WebSocket('ws://localhost:8080/api/ws');
// Ensure session cookie is sent with the connection
```

## 🔒 Security Features

1. **Password Security**: bcrypt hashing with configurable cost
2. **Session Management**: Secure, HTTP-only cookies with TTL
3. **Rate Limiting**: Configurable request limits per IP
4. **CORS**: Configurable allowed origins
5. **WebSocket Auth**: Cookie-based authentication for WS connections
6. **Input Validation**: Request validation and sanitization

## 📨 Message Queue Events

The system publishes events for key actions:

- `USER_LOGGED_IN`: Published when a user successfully logs in
- `USER_REGISTERED`: Published when a new user registers
- `USER_LOGGED_OUT`: Published when a user logs out

Events are consumed by the worker process defined in `main.go`. Add custom event handlers in the `handleEvent` function.

## 🔧 Development

### Adding New Features

1. **New Models**: Add to `internal/models/`
2. **Business Logic**: Add to appropriate service in `internal/`
3. **HTTP Handlers**: Add to `internal/http/handlers/`
4. **Routes**: Register in `internal/http/router.go`
5. **Middleware**: Add to `internal/http/middleware/`

### Database Migrations

Create new migration files:
```bash
# Up migration
migrations/003_your_migration.up.sql

# Down migration
migrations/003_your_migration.down.sql
```

## 📊 Architecture Decisions

### Session Storage
- **Primary**: Redis for performance and automatic expiration
- **Fallback**: PostgreSQL sessions table (optional, not currently used)
- Sessions auto-refresh when 1 hour from expiration

### Message Queue
- Uses Redis Streams for simplicity
- Can be swapped for RabbitMQ by implementing MQ interface
- Consumer groups for reliable message processing

### WebSocket Authentication
- Reuses HTTP session cookies
- No separate WS authentication required
- Automatically disconnects on session expiration

## 🚀 Production Considerations

1. **Environment Variables**
   - Generate a strong `SESSION_SECRET` (64+ characters)
   - Set `COOKIE_SECURE=true`
   - Set `ENV=production`

2. **Database**
   - Enable SSL (`DB_SSLMODE=require`)
   - Use connection pooling
   - Regular backups

3. **Redis**
   - Enable persistence (RDB/AOF)
   - Set password authentication
   - Consider Redis Cluster for scale

4. **Security**
   - Use HTTPS in production
   - Configure proper CORS origins
   - Implement API rate limiting
   - Add request logging and monitoring

5. **Performance**
   - Adjust database connection pools
   - Monitor Redis memory usage
   - Load test WebSocket connections
   - Implement caching strategies

## 🤝 Contributing

This is a foundational backend. To extend:
1. Add delivery-specific business logic
2. Implement driver matching algorithms
3. Add pricing and payment systems
4. Implement real-time tracking features

## 📝 License

MIT

## 👥 Authors

Created as a production-ready backend foundation for delivery systems.

---

**Note**: This is a foundation/skeleton. Business logic for deliveries, matching, pricing, and tracking should be built on top of this authentication and infrastructure layer.
