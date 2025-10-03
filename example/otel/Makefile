.PHONY: help deps run test load stop clean logs ui

# Default target
help:
	@echo "Hyperion OTel Example - Available targets:"
	@echo "  make deps        - Download Go dependencies"
	@echo "  make hyperdx-up  - Start HyperDX local instance"
	@echo "  make hyperdx-down- Stop HyperDX"
	@echo "  make hyperdx-logs- View HyperDX logs"
	@echo "  make run         - Run the example application"
	@echo "  make test        - Run tests"
	@echo "  make load        - Generate sample traffic"
	@echo "  make ui          - Open HyperDX UI in browser"
	@echo "  make clean       - Clean up everything"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Start HyperDX
hyperdx-up:
	@echo "Starting HyperDX local instance..."
	docker-compose up -d
	@echo "Waiting for HyperDX to be ready (this may take a minute)..."
	@sleep 10
	@echo "HyperDX is starting up!"
	@echo "  - UI: http://localhost:8080"
	@echo "  - OTLP HTTP: http://localhost:4318"
	@echo "  - OTLP gRPC: localhost:4317"

# Stop HyperDX
hyperdx-down:
	@echo "Stopping HyperDX..."
	docker-compose down

# View HyperDX logs
hyperdx-logs:
	docker-compose logs -f hyperdx

# Run the example application
run: deps
	@echo "Running example application..."
	@echo "Server will be available at http://localhost:8090"
	go run cmd/app/main.go

# Run tests
test:
	go test -v ./...

# Generate sample traffic
load:
	@echo "Generating sample traffic..."
	@for i in {1..10}; do \
		curl -s http://localhost:8090/api/users/$$i > /dev/null && echo "Request $$i: GET /api/users/$$i"; \
		sleep 0.5; \
	done
	@echo ""
	@for i in {1..5}; do \
		curl -s http://localhost:8090/api/slow > /dev/null && echo "Request: GET /api/slow"; \
		sleep 1; \
	done
	@echo ""
	@for i in {1..3}; do \
		curl -s http://localhost:8090/api/error > /dev/null && echo "Request: GET /api/error (expected error)"; \
		sleep 0.5; \
	done
	@echo ""
	@echo "Sample traffic generation completed!"
	@echo "Check HyperDX UI at http://localhost:8080"

# Open HyperDX UI
ui:
	@echo "Opening HyperDX UI..."
	@open http://localhost:8080 || xdg-open http://localhost:8080 || echo "Please open http://localhost:8080 in your browser"

# Clean up
clean: hyperdx-down
	@echo "Cleaning up..."
	rm -f go.sum
	docker-compose down -v
	@echo "Cleanup completed!"

# Full demo workflow
demo: hyperdx-up
	@echo "Waiting for HyperDX to fully start..."
	@sleep 30
	@echo ""
	@echo "Starting demo application..."
	@make run &
	@echo "Waiting for application to start..."
	@sleep 5
	@echo ""
	@echo "Generating sample traffic..."
	@make load
	@echo ""
	@echo "Demo completed! Check HyperDX UI at http://localhost:8080"
