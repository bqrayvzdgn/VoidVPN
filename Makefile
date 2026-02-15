BINARY_NAME=voidvpn
MODULE=github.com/voidvpn/voidvpn
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w \
	-X $(MODULE)/pkg/version.Version=$(VERSION) \
	-X $(MODULE)/pkg/version.Commit=$(COMMIT) \
	-X $(MODULE)/pkg/version.BuildDate=$(BUILD_DATE)"

.PHONY: all build build-windows build-linux build-darwin clean test vet fmt install

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/voidvpn/

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME).exe ./cmd/voidvpn/

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux ./cmd/voidvpn/

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin ./cmd/voidvpn/

build-all: build-windows build-linux build-darwin

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-linux $(BINARY_NAME)-darwin
	rm -f coverage.out

test:
	go test ./... -v -count=1 -timeout 120s

test-coverage:
	go test ./... -coverprofile=coverage.out -timeout 120s
	go tool cover -func=coverage.out

vet:
	go vet ./...

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./... || echo "golangci-lint not installed"

install:
	go install $(LDFLAGS) ./cmd/voidvpn/

docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t voidvpn:$(VERSION) .

completions:
	mkdir -p completions
	./$(BINARY_NAME) completion bash > completions/voidvpn.bash
	./$(BINARY_NAME) completion zsh > completions/_voidvpn
	./$(BINARY_NAME) completion fish > completions/voidvpn.fish
	./$(BINARY_NAME) completion powershell > completions/voidvpn.ps1
