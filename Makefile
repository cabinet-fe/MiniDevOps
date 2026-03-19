.PHONY: dev dev-backend dev-frontend build clean

# Development
dev:
	@trap 'kill 0' INT TERM; \
	mkdir -p cmd/server/dist && touch cmd/server/dist/.gitkeep; \
	(cd cmd/server && go run . --config ../../config.yaml) & \
	(cd web && bun run dev) & \
	wait

dev-backend:
	@mkdir -p cmd/server/dist && touch cmd/server/dist/.gitkeep
	cd cmd/server && go run . --config ../../config.yaml

dev-frontend:
	cd web && bun run dev

# Production build
build: build-frontend build-backend

build-frontend:
	cd web && bun run build
	rm -rf cmd/server/dist
	cp -r web/dist cmd/server/dist

build-backend:
	CGO_ENABLED=1 go build -o buildflow ./cmd/server

# Cross-compile
build-linux:
	cd web && bun run build
	rm -rf cmd/server/dist
	cp -r web/dist cmd/server/dist
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o buildflow-linux ./cmd/server

# Clean
clean:
	rm -rf buildflow buildflow-linux cmd/server/dist web/dist data/
