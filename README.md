# Balloon Popping Game

An interactive demo designed to generate streaming data for real-time analytics demonstrations. The game uses WebSocket for real-time interactions and streams events to Kafka, making it perfect for showcasing streaming databases and real-time analytics capabilities.

> [!NOTE]
> This README provides comprehensive setup instructions to run the Balloon Popping Game demo locally.

## 🚀 Features & Use Cases

- Real-time player performance analytics
- Time-series analysis of game events
- Pattern detection in player behavior
- Color popularity and bonus effectiveness tracking
- Visual effects for bonus balloons
- Real-time score updates with animations

---

## ⚡️ Prerequisites

> [!IMPORTANT]
> **Required Tools**
>
> - [Go](https://go.dev/dl/) >= 1.23.4
> - [Docker](https://docs.docker.com/get-docker/) & Docker Compose
> - [HTTPie](https://httpie.io/cli) - for API testing
> - [jq](https://jqlang.github.io/jq/) - for JSON processing
> - [Task](https://taskfile.dev/) (optional) - for running tasks
> - [OpenSSL](https://www.openssl.org/) - for key verification
> - Web browser with HTML5 support
> - [Kafka](https://kafka.apache.org/downloads) (local or Docker)

> [!TIP]
> For complete infrastructure setup, follow: [Balloon Popper Demo Infrastructure Guide](https://kameshsampath.github.io/balloon-popper-demo/)

---

## 🛠️ Setup & Configuration

### Step 1: Generate JWT Keys

Create the RSA key pair for JWT token signing:

```shell
go run cmd/main.go jwt-keys
```

**Using Taskfile (requires environment variables):**

```shell
task jwt-keys
```

> [!NOTE]
> The Taskfile version requires `JWT_KEY_SECRET_NAME` environment variable to be set.

### Step 2: Create Admin User

Create an admin user for game management:

```shell
go run cmd/main.go user -u admin -p password1234! -r admin -e 'balloon-game-admin@example.com'
```

**Using Taskfile (requires environment variables):**

```shell
task create-admin
```

> [!NOTE]
> The Taskfile version requires `BALLOON_POPPER_ADMIN_PASSWORD` environment variable to be set.

### Step 3: Verify RSA Keys (Optional)

Verify that the JWT keys were created correctly:

```shell
# Basic verification
openssl rsa -in keys/jwt-private-key -inform PEM -passin file:keys/.pass

# Detailed key information
openssl rsa -in keys/jwt-private-key -inform PEM -passin file:keys/.pass -text -noout
```

### Step 4: Configuration File

Create `config.json` in the project root:

```json
{
  "dev": {
    "host": "localhost",
    "port": "8080",
    "username": "admin",
    "password": "password1234!"
  }
}
```

> [!WARNING]
> **Security Notice**
>
> - Replace `password1234!` with your actual admin password
> - Do not commit `config.json` to version control
> - Consider using environment variables for production

---

## 🏃‍♂️ Running the Application

### Start the Server

```shell
go run cmd/main.go server -k ./keys/jwt-private-key -p $(cat ./keys/.pass) -c ./config/users.json
```

**Using Taskfile:**

```shell
task server
```

The server will start on `http://localhost:8080`

---

## 🎮 Game Management API

### Option 1: Using HTTPie Commands

**Login and get JWT token:**

```shell
http --form POST localhost:8080/login \
  username=admin \
  password='password1234!' \
  Accept:application/json
```

**Start the game (replace `<TOKEN>` with JWT from login):**

```shell
http POST localhost:8080/admin/start \
  Authorization:"Bearer <TOKEN>" \
  Content-Type:application/json \
  Accept:application/json \
  '{}'
```

**Stop the game:**

```shell
http POST localhost:8080/admin/stop \
  Authorization:"Bearer <TOKEN>" \
  Content-Type:application/json \
  Accept:application/json \
  '{}'
```

### Option 2: Using Provided Scripts

**Start the game:**

```shell
./start_game.sh
```

**Stop the game:**

```shell
./stop_game.sh
```

> [!TIP]
> The scripts automatically handle login and token extraction using your `config.json` settings.

---

## 🖥️ Playing the Game

1. **Start the server** (see above)
2. **Start the game** using API or scripts
3. **Open the game UI** in your browser: <http://localhost:8080>

---

## 📚 API Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/login` | Authenticate user | No |
| POST | `/admin/start` | Start game | Yes (Bearer token) |
| POST | `/admin/stop` | Stop game | Yes (Bearer token) |
| GET | `/health` | Health check | No |

---

## 🔧 Troubleshooting

> [!WARNING]
> **Common Issues**
>
> - **"missing or malformed JWT"**: Ensure you're sending the `Authorization: Bearer <token>` header
> - **Login fails**: Check username/password in `config.json` matches created user
> - **Key errors**: Verify JWT keys exist in `keys/` directory
> - **Port conflicts**: Ensure port 8080 is available

---

## 📖 Related Documentation

- [Echo Framework](https://echo.labstack.com/) - Web framework used
- [Kafka Documentation](https://kafka.apache.org/documentation/) - Message streaming
- [HTML5 Canvas Tutorial](https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API/Tutorial) - Game graphics
- [JWT.io](https://jwt.io/) - JSON Web Tokens
- [Taskfile](https://taskfile.dev/) - Task runner

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
