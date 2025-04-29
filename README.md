# Chatting Service

This repository contains a **BES Chatting** implementation, complete with monitoring and observability using **Prometheus** and **Grafana**, and automated orchestration via **Docker Compose** and **Makefile**.

---

## 🚀 Features

- **OpenAI GPT Integration**: Communicate with ChatGPT for natural language conversations.
- **Configurable via `config/config.yml`**: Easily set API keys, ports, and service endpoints.
- **Dockerized Services**:
  - **Chatbot API** (Golang/Huma)
  - **Prometheus** for metrics scraping
  - **Grafana** for dashboarding
- **Makefile Automation**: One-command builds, starts, and stops all services.
- **Health Checks**: Built-in endpoints and Prometheus probes.
- **Scalable**: Horizontal scaling via Docker Compose replicas.

---

## 📋 Prerequisites

- **Docker** ≥ 20.10
- **Docker Compose** ≥ 1.29
- **Make** utility
- **OpenAI API Key**

---

## 🔧 Configuration

1. Copy the example config file and customize:
   ```bash
   cp config/config.yml.example config/config.yml
   ```
2. Open `config/config.yml` and set your variables:
   ```yaml
   server:
    port: 
    mode: 
    log_level: 
    log_file: "./tmp/console.log"
    grateful_shutdown:
      shutdown_time: 
      read_timeout: 
      write_timeout: 
      idle_timeout: 

    api:
      api_service_name: chatbot_gpt
      api_version: v1.0.0
      tls_cert_path: "./cert/ws_cert.pem"
      tls_key_path: "./cert/ws_key.pem"

    pkg:
      openai:
        token: 
        models:

      redis:
        dsn: 
   ```

---

## 🐳 Docker Compose Setup

This project uses three separate Docker Compose files:

- **`docker-compose.yml`**: Defines the **Chatbot API** service.
- **`docker-compose.prometheus.yml`**: Defines the **Prometheus** monitoring service.
- **`docker-compose.grafana.yml`**: Defines the **Grafana** dashboard service.

You can run all services together with:
```bash
docker-compose \
  -f docker-compose.yml \
  -f docker-compose.prometheus.yml \
  -f docker-compose.grafana.yml \
  up -d
```

To stop and remove all services:
```bash
docker-compose \
  -f docker-compose.yml \
  -f docker-compose.prometheus.yml \
  -f docker-compose.grafana.yml \
  down
```

## 🛠️ Makefile

```makefile
# Makefile commands for development and deployment

# Build images
build:
	docker-compose build

# Start all services (chatbot, Prometheus, Grafana)
up:
	docker-compose up -d

# Stop all services
down:
	docker-compose down

# Tail logs
logs:
	docker-compose logs -f

# Run health check
health:
	curl -f http://localhost:${CHATBOT_PORT}/health || echo "BES Chatting is down!"

# Rebuild and restart
restart: build down up
```

---

## 📖 Usage

1. Start services:
   ```bash
   make up
   ```
2. Open chatbot API at `http://localhost:8000/docs` (Swagger UI) or call gRPC endpoint.
3. View Prometheus at `http://localhost:9090`.
4. View Grafana at `http://localhost:3000` (default admin/admin).

---

## 📝 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

