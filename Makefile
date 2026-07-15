.PHONY: dev dev-backend dev-frontend build build-frontend build-backend build-linux build-linux-arm64 build-win build-agent-linux build-agent-win openapi-projection openapi-check clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
FRONTEND_DIR ?= web-v2
export VITE_APP_VERSION := $(VERSION)
# Dev encryption key for web-v2 password_cipher (must match config.yaml encryption.key)
export VITE_BEDROCK_ENCRYPTION_KEY ?= 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
LDFLAGS := -s -w -X main.version=$(VERSION)

dev:
	@trap 'kill 0' INT TERM; \
	(cd cmd/server && go run -tags dev . --config ../../config.yaml) & \
	(cd $(FRONTEND_DIR) && vp dev) & \
	wait

dev-backend:
	cd cmd/server && go run -tags dev . --config ../../config.yaml

dev-frontend:
	cd $(FRONTEND_DIR) && vp dev

build-frontend:
	cd $(FRONTEND_DIR) && vp install && vp build

build-backend:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bedrock ./cmd/server

build: build-frontend
	rm -rf cmd/server/dist && cp -r $(FRONTEND_DIR)/dist cmd/server/dist
	$(MAKE) build-backend

build-linux: build-frontend
	rm -rf cmd/server/dist && cp -r $(FRONTEND_DIR)/dist cmd/server/dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bedrock-linux-amd64 ./cmd/server

build-linux-arm64: build-frontend
	rm -rf cmd/server/dist && cp -r $(FRONTEND_DIR)/dist cmd/server/dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bedrock-linux-arm64 ./cmd/server

build-win: build-frontend
	rm -rf cmd/server/dist && cp -r $(FRONTEND_DIR)/dist cmd/server/dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bedrock-windows-amd64.exe ./cmd/server

build-agent-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bedrock-agent-linux-amd64 ./cmd/agent

build-agent-win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bedrock-agent-windows-amd64.exe ./cmd/agent

openapi-projection:
	go run ./tools/openapi-project api/openapi.yaml api/openapi.3.1.projection.yaml

openapi-check: openapi-projection
	@git diff --exit-code -- api/openapi.3.1.projection.yaml || \
		(echo "openapi.3.1.projection.yaml is out of date; run make openapi-projection" >&2; exit 1)

clean:
	rm -rf bedrock* cmd/server/dist $(FRONTEND_DIR)/dist web/dist data/
