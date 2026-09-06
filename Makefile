APP_NAME := latinabot
MAIN_PACKAGE := ./cmd/bot

.PHONY: all build run test clean tidy docker-build

all: test build

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(APP_NAME) $(MAIN_PACKAGE)

run:
	go run $(MAIN_PACKAGE)

test:
	go test -v -race ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ temp/

docker-build:
	docker build -t $(APP_NAME):latest -f Dockerfile ..
