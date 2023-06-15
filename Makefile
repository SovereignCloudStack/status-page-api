APP_NAME=status-page-api
BIN_DIR=bin
DOC_DIR=docs

.PHONY: all go-fmt go-fump go-gci go-format go-lint go-build go-doc clean serve db-create db-start db-stop db-remove db-restart

all: go-format go-lint go-test go-build

go-fmt:
	go fmt ./...

go-fump:
	gofumpt -w .

go-gci:
	gci write --skip-generated -s standard -s default .

go-format: fmt fump gci

go-lint:
	golangci-lint run

go-test:
	go test ./...

$(BIN_DIR):
	@mkdir -p $@

go-build: $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) main.go

${DOC_DIR}:
	mkdir -p ${DOC_DIR}

go-doc: ${DOC_DIR}
	gomarkdoc --output '${DOC_DIR}/{{.Dir}}/README.md' ./...

clean:
	go clean
	rm -f $(BIN_DIR)/*
	rm -f $(DOC_DIR)/*

serve: go-build
	$(BIN_DIR)/$(APP_NAME)

db-create:
	docker create -p 5432:5432 -e POSTGRES_PASSWORD=debug -e POSTGRES_USER=postgres -e POSTGRES_DB=postgres --name scs-status-page-api-db postgres:latest

db-start:
	docker start scs-status-page-api-db

db-stop:
	docker stop scs-status-page-api-db

db-remove:
	docker container rm scs-status-page-api-db

db-restart: db-stop db-remove db-create db-start
