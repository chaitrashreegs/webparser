# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOFMT := $(GOCMD) fmt
GOLINT := golint

# Docker parameters
DOCKER_BUILD_CMD := docker build
DOCKER_RUN_CMD := docker run
DOCKER_PUSH_CMD := docker push

# Application parameters
APP_NAME := myapp
APP_VERSION := 1.0.0
DOCKER_REPO := myrepo/myapp
DOCKER_IMAGE := $(DOCKER_REPO):$(APP_VERSION)

.PHONY: all fmt lint build run docker-build docker-run docker-push clean

all: build

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

lint:
	@echo "Linting code..."
	$(GOLINT) ./...

build:
	@echo "Building application..."
	$(GOBUILD) -o $(APP_NAME) ./cmd/main.go

run:
	@echo "Running application..."
	./$(APP_NAME)

docker-build: build
	@echo "Building Docker image..."
	$(DOCKER_BUILD_CMD) -t $(DOCKER_IMAGE) .

docker-run: docker-build
	@echo "Running Docker container..."
	$(DOCKER_RUN_CMD) -v /tmp/test:/tmp/test -p 8090:8090 $(DOCKER_IMAGE) --file-path=/tmp/test

docker-push: docker-build
	@echo "Pushing Docker image to repository..."
	$(DOCKER_PUSH_CMD) $(DOCKER_IMAGE)

clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)
