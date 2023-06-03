# Should be adapted to the project name as it declared in go.mod.
PRJ=$(shell awk '{print $2}' go.mod)

# Optionally set these args as environment vars in the shell. You
# could also pass them as parameters of `make`.
# For example: make build CMD=console
CMD?=xcom
FLAGS?=-v
CLEANUP?=
GITHUB_TOKEN?=
REGISTRY?=docker.com
RACE_DETECTOR?=false

default: lint test

# Optional includes that depend on project:
-include doc.mk

init:
	#go mod init 0xADE/pad
	fyne-cross linux --pull

cross:
	fyne-cross linux -arch amd64 -tags wayland,egl ./cmd/$(CMD)

build:
	CGO_ENABLED=1 go build $(FLAGS) $(RACE) ./cmd/$(CMD)

build-all:
	$(foreach dir,$(wildcard cmd/*), go build $(FLAGS) ./$(dir);)

vendor:
	go mod download && go mod tidy && go mod vendor

docker:
	DOCKER_BUILDKIT=1 docker build -f Dockerfile -t cryptoboy.io/$(PRJ) \
	--build-arg GITHUB_TOKEN=$(GITHUB_TOKEN) \
	--build-arg REGISTRY=$(REGISTRY) \
	--build-arg CGO=$(CGO) \
	--build-arg RACE=$(RACE) .

lint:
	golangci-lint run -v ./...

test:
	go test $(FLAGS) ./...

mod:
	go mod tidy

load-tools:
	@go install -v github.com/go-task/task/v3/cmd/task@latest

generate:
	go generate $(FLAGS) ./...

clean:
	@echo $(CLEANUP)
	$(foreach f,$(CLEANUP),rm -rf $(f);)

.PHONY: build build-all vendor docker lint test tidy generate clean
