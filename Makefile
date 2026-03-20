.PHONY: dev build-linux build-win clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

dev:
	@trap 'kill 0' INT TERM; \
	(cd cmd/server && go run -tags dev . --config ../../config.yaml) & \
	(cd web && bun run dev) & \
	wait

build-linux:
	cd web && bun install && bun run build
	rm -rf cmd/server/dist && cp -r web/dist cmd/server/dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o buildflow-linux-amd64 ./cmd/server

build-win:
	cd web && bun install && bun run build
	rm -rf cmd/server/dist && cp -r web/dist cmd/server/dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o buildflow-windows-amd64.exe ./cmd/server

clean:
	rm -rf buildflow* cmd/server/dist web/dist data/
