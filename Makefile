APP_NAME=status-page-api
BIN_DIR=bin
CMD_DIR=cmd
DOC_DIR=docs
HASH=$(shell git rev-parse --short HEAD)

CONTAINER_RUNTIME := $(shell command -v docker 2> /dev/null)
ifndef CONTAINER_RUNTIME
CONTAINER_RUNTIME := $(shell command -v podman 2> /dev/null)
endif

SHELL := /bin/bash

.PHONY: all go-fmt go-fump go-gci go-format go-lint go-build go-doc clean serve db-create db-start db-stop db-remove db-restart guard-container-cmd

all: go-format go-lint go-test go-build

go-fmt:
	go fmt ./...

go-fump:
	gofumpt -w .

go-gci:
	gci write --skip-generated -s standard -s default .

go-format: go-fmt go-fump go-gci

go-lint:
	golangci-lint run

go-lint-fix:
	# try auto-fix for lint errors
	golangci-lint run --fix

go-test:
	go test ./...

go-test-coverage:
	go test -coverprofile coverage.out ./...
	go tool cover -func coverage.out
	rm -f coverage.out

$(BIN_DIR):
	@mkdir -p $@

go-build: $(BIN_DIR)
	go build -ldflags "-s -w" -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)/$(APP_NAME)/main.go

${DOC_DIR}:
	mkdir -p ${DOC_DIR}

go-doc: ${DOC_DIR}
	gomarkdoc --output '${DOC_DIR}/{{.Dir}}/README.md' ./...

clean:
	go clean
	rm -f $(BIN_DIR)/*
	rm -f $(DOC_DIR)/*

serve: go-build
	source ./load-secrets.sh && ./$(BIN_DIR)/$(APP_NAME)

db-create: guard-container-cmd
	${CONTAINER_RUNTIME} create -p 5432:5432 -e POSTGRES_PASSWORD=debug -e POSTGRES_USER=postgres -e POSTGRES_DB=postgres --name scs-${APP_NAME}-db docker.io/library/postgres:latest

db-start: guard-container-cmd
	${CONTAINER_RUNTIME} start scs-${APP_NAME}-db

db-stop: guard-container-cmd
	${CONTAINER_RUNTIME} stop scs-${APP_NAME}-db

db-remove: guard-container-cmd
	${CONTAINER_RUNTIME} container rm scs-${APP_NAME}-db

db-restart: db-stop db-remove db-create db-start

container-build: guard-container-cmd
	${CONTAINER_RUNTIME} build -t ${APP_NAME}:latest -t ${APP_NAME}:${HASH} -f Containerfile .

container-push-harbor: guard-container-cmd
	${CONTAINER_RUNTIME} tag ${APP_NAME}:${HASH} registry.scs.community/status-page/${APP_NAME}:${HASH}
	${CONTAINER_RUNTIME} tag ${APP_NAME}:latest registry.scs.community/status-page/${APP_NAME}:latest

	${CONTAINER_RUNTIME} push registry.scs.community/status-page/${APP_NAME}:${HASH}
	${CONTAINER_RUNTIME} push registry.scs.community/status-page/${APP_NAME}:latest

# Check, if the CONTAINER_RUNTIME variable is set
guard-container-cmd:
	@if [ "${CONTAINER_RUNTIME}" = "" ]; then \
		echo "Either DOCKER or PODMAN must be installed"; \
		exit 1; \
	fi
