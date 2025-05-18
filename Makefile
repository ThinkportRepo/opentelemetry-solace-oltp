.PHONY: build start debug docker-build docker-build-local docker-push version-major version-minor version-patch

# Include environment variables from .env file if it exists
-include .env
export

# Version management
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")

version-major:
	@echo "Current version: $(CURRENT_VERSION)"
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{print $$1+1".0.0"}'); \
	echo "New version: $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push --tags; \
	echo "Created and pushed tag v$$NEW_VERSION"

version-minor:
	@echo "Current version: $(CURRENT_VERSION)"
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{print $$1"."$$2+1".0"}'); \
	echo "New version: $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push --tags; \
	echo "Created and pushed tag v$$NEW_VERSION"

version-patch:
	@echo "Current version: $(CURRENT_VERSION)"
	@NEW_VERSION=$$(echo $(CURRENT_VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}'); \
	echo "New version: $$NEW_VERSION"; \
	git tag -a "v$$NEW_VERSION" -m "Release v$$NEW_VERSION"; \
	git push --tags; \
	echo "Created and pushed tag v$$NEW_VERSION"

# Build the OpenTelemetry Collector
build:
	ocb --config builder-config.yaml

# Start the OpenTelemetry Collector
start:
	./otelcol-dev/otelcol-dev --config collector-config.yaml

# Build and start the OpenTelemetry Collector
rebuild:
	make build
	make start

# Debug the OpenTelemetry Collector
debug:
	./otelcol-dev/otelcol-dev --config collector-config.yaml --log-level=debug

# Build Docker image for Linux AMD64 (default)
docker-build:
	docker build -t ghcr.io/thinkportrepo/opentelemetry-receiver-solace:latest .

# Build Docker image for Mac ARM64 (local development)
docker-build-local:
	docker build -t ghcr.io/thinkportrepo/opentelemetry-receiver-solace:local .

docker-push:
	docker push ghcr.io/thinkportrepo/opentelemetry-receiver-solace:latest 