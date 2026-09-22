.PHONY: all relay client test clean

all: relay client

relay:
	@echo "==> Building Lagvex Relay (Linux amd64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/lagvex-relay ./cmd/lagvex-relay

relay-arm64:
	@echo "==> Building Lagvex Relay (Linux arm64)..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/lagvex-relay-arm64 ./cmd/lagvex-relay

client:
	@echo "==> Building Lagvex Client (Windows amd64)..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -H=windowsgui" -o bin/lagvex-client.exe ./cmd/lagvex-client

client-mac:
	@echo "==> Building Lagvex Client (macOS arm64)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/lagvex-client-darwin-arm64 ./cmd/lagvex-client

client-linux:
	@echo "==> Building Lagvex Client (Linux / Steam Deck amd64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/lagvex-client-linux-amd64 ./cmd/lagvex-client

test:
	@echo "==> Running All Unit Tests..."
	go test -v ./...

clean:
	rm -rf bin/lagvex-relay bin/lagvex-relay-arm64 bin/lagvex-client.exe
