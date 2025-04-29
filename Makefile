DOCKER_BIN := docker
COMPOSE_BIN := docker compose
COMPOSE_TOOL_RUN := $(COMPOSE_BIN) run --rm --service-ports tool

mod:
	go mod tidy
	go mod vendor

init: redis

build:
	-@$(DOCKER_BIN) image rm -f besanh/chatbot_gpt:latest
	@$(DOCKER_BIN) build -t besanh/chatbot_gpt:latest .
	-@$(DOCKER_BIN) rmi ${docker images -f "dangling=true" -q}

test:
	@$(COMPOSE_TOOL_RUN) sh -c 'go test -mod=vendor -vet=all -coverprofile=coverage.out -failfast -timeout 5m ./...'

redis:
	@$(COMPOSE_BIN) up redis -d

migrate:
	@$(COMPOSE_TOOL_RUN) sh -c 'migrate -path ./data/migration -database $$POSTGRES_URL up'

redo-migrate:
	@$(COMPOSE_TOOL_RUN) sh -c 'migrate -path ./data/migration -database $$POSTGRES_URL drop'
	@$(COMPOSE_TOOL_RUN) sh -c 'migrate -path ./data/migration -database $$POSTGRES_URL up'

gen-pem:
	@$(COMPOSE_TOOL_RUN) sh -c "openssl genpkey -algorithm RSA -out ./cert/ws_key.pem"
	@$(COMPOSE_TOOL_RUN) sh -c "openssl req -x509 -new -nodes -key ./cert/ws_key.pem -sha256 -days 365 -out ./cert/ws_cert.pem"
