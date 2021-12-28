.PHONY: all build docker build_docker clean createdirs lint lint_pedantic mrproper run run_mod_off

TARGET := matchmaker-http-api
BIN_DIR := bin
TESTS_DIR := $(BIN_DIR)/tests

all: createdirs build build_docker push_docker

build:
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(TARGET) -v .

run_mod_off:
	GO111MODULE=off go run main.go -c .env

run:
	go run main.go -c .env

docker: createdirs build build_docker push_docker

build_docker:
	docker build  . -t hub.docker.com/djumanoff/$(TARGET):$(v)

push_docker:
	docker push hub.docker.com/djumanoff/$(TARGET):$(v)

run_docker:
	docker run --env-file=.env hub.docker.com/djumanoff/$(TARGET):$(v)

clean:
	@rm $(BIN_DIR)/$(TARGET)

createdirs:
	@mkdir -p $(BIN_DIR)
