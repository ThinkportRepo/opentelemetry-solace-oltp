.PHONY: build start debug docker-build docker-push version-major version-minor version-patch otel-test-spans stop check help kill test

# Colors for output
BLUE := \033[0;34m
BOLD := \033[1m
GREEN := \033[0;32m
RED := \033[0;31m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Include environment variables from .env file if it exists
-include .env
export

# Version management
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")

help:
	@echo "\n${BOLD}${BLUE}Thinkport GmbH${NC}"
	@echo "\n${BLUE}Available Tasks:${NC}\n"
	@echo "${GREEN}build${NC}              - Build OpenTelemetry Collector"
	@echo "${GREEN}start${NC}              - Start OpenTelemetry Collector"
	@echo "${GREEN}debug${NC}              - Start OpenTelemetry Collector in debug mode"
	@echo "${GREEN}stop${NC}               - Stop OpenTelemetry Collector"
	@echo "${GREEN}check${NC}              - Check if required ports are available"
	@echo "${GREEN}kill${NC}               - Kill processes on ports 4317 and 4318"
	@echo "${GREEN}docker-build${NC}       - Build Docker image for Linux AMD64"
	@echo "${GREEN}docker-push${NC}        - Push Docker image"
	@echo "${GREEN}otel-test-spans${NC}    - Send test spans"
	@echo "${GREEN}test${NC}               - Run Go-Tests"
	@echo "${GREEN}version-major${NC}      - Increment major version"
	@echo "${GREEN}version-minor${NC}      - Increment minor version"
	@echo "${GREEN}version-patch${NC}      - Increment patch version"

version-major:
	@echo "${BLUE}Current version:${NC} $(CURRENT_VERSION)"
	@git fetch --tags --force
	@git tag -l | xargs git tag -d
	@git fetch --tags
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{print $$1+1".0.0"}'); \
	echo "${GREEN}New version:${NC} $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push --tags; \
	echo "${GREEN}Tag v$$NEW_VERSION created and pushed${NC}"

version-minor:
	@echo "${BLUE}Current version:${NC} $(CURRENT_VERSION)"
	@git fetch --tags --force
	@git tag -l | xargs git tag -d
	@git fetch --tags
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{print $$1"."$$2+1".0"}'); \
	echo "${GREEN}New version:${NC} $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push --tags; \
	echo "${GREEN}Tag v$$NEW_VERSION created and pushed${NC}"

version-patch:
	@echo "${BLUE}Current version:${NC} $(CURRENT_VERSION)"
	@git fetch --tags --force
	@git tag -l | xargs git tag -d
	@git fetch --tags
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}'); \
	echo "${GREEN}New version:${NC} $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push --tags; \
	echo "${GREEN}Tag v$$NEW_VERSION created and pushed${NC}"

# Build the OpenTelemetry Collector
build:
	@echo "${BLUE}Building OpenTelemetry Collector … ${NC}"
	@ocb --config collector/builder-config.yaml
	@echo "${GREEN}Build completed${NC}"

# Start the OpenTelemetry Collector
start:
	@echo "${BLUE}Starting OpenTelemetry Collector … ${NC}"
	@SOLACE_TRUST_STORE_PATH=truststore SOLACE_HOST=$(SOLACE_HOST) SOLACE_QUEUE=$(SOLACE_QUEUE) SOLACE_USERNAME=$(SOLACE_USERNAME) SOLACE_PASSWORD=$(SOLACE_PASSWORD) SOLACE_VPN=$(SOLACE_VPN) DD_SITE=$(DD_SITE) DD_API_KEY=$(DD_API_KEY) ./otelcol-dev/otelcol-dev --config collector/collector-config.yaml

# Build and start the OpenTelemetry Collector
rebuild:
	make stop
	make build
	make start

# Build Docker image for Linux AMD64 (default)
docker-build:
	@echo "${BLUE}Building Docker image for Linux AMD64 … ${NC}"
	@docker build -t ghcr.io/thinkportrepo/opentelemetry-receiver-solace:latest -f collector/Dockerfile .
	@echo "${GREEN}Docker build completed${NC}"

# Push Docker image to GitHub Container Registry
docker-push:
	@echo "${BLUE}Pushing Docker image…${NC}"
	@docker push ghcr.io/thinkportrepo/opentelemetry-receiver-solace:latest
	@echo "${GREEN}Docker push completed${NC}"

# Send test spans using otel-cli
otel-test-spans:
	@echo "${BLUE}Starting test span generator … ${NC}"
	@while true; do \
		printf "${YELLOW}Sending test span …${NC}"; \
		otel-cli span \
			--service "test-service" \
			--name "test-span" \
			--endpoint "0.0.0.0:4317" \
			--protocol grpc \
			--insecure \
			--kind client \
			--attrs "test.attribute=value"; \
		printf "${GREEN}sent${NC}\n"; \
		sleep 1; \
	done

# Stop the OpenTelemetry Collector
stop:
	@printf "${BLUE}Stopping OpenTelemetry Collector … ${NC} "
	@pkill -f "otelcol-dev" || true
	@echo "${GREEN}stopped${NC}"

# Check if required ports are available
check:
	@printf "${BLUE}Checking port availability … ${NC} "
	@if lsof -i :4317 >/dev/null 2>&1; then \
		echo "${RED}Error: Port 4317 (OTLP/gRPC) is already in use${NC}"; \
		exit 1; \
	fi
	@if lsof -i :8888 >/dev/null 2>&1; then \
		echo "${RED}Error: Port 8888 (Prometheus) is already in use${NC}"; \
		exit 1; \
	fi
	@echo "${GREEN}OK - All required ports are available${NC}"

kill:
	@echo "Searching and killing processes on ports 4317 and 4318..."
	-lsof -ti :4317 | xargs -r kill -9
	-lsof -ti :4318 | xargs -r kill -9
	@echo "Done."

test:
	@echo "${BLUE}Running Go-Tests … ${NC}"
	@go test ./... -v
	@echo "${GREEN}Tests abgeschlossen${NC}"
