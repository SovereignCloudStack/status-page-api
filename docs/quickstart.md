# Quickstart

See [requirements](./requirements.md)

## Run as container

The quickest way to start working with the API server, is to run the container directly.

```bash
docker run --rm --network host -e STATUS_PAGE_DATABASE_CONNECTION_STRING="host=localhost user=postgres dbname=postgres port=5432 password=debug sslmode=disable" -e STATUS_PAGE_VERBOSE=3 registry.scs.community/status-page/status-page-api:latest
```

## Compile and run the binary

Compiling the binary and running it is equally easy.

```bash
go build -o /bin/status-page-api cmd/status-page-api/main.go

STATUS_PAGE_DATABASE_CONNECTION_STRING="host=localhost user=postgres dbname=postgres port=5432 password=debug sslmode=disable" STATUS_PAGE_VERBOSE=3 ./bin/status-page-api
```

## Running tests

The status page API server tests it's API handler and database code with a plethora of tests. These tests can be run with `go`.

```bash
go test ./...
```

Furthermore test can be run to create a coverage profile.

```bash
go test -coverprofile coverage.out ./...
```

This cover profile can be used to analyze code coverage

```bash
# per function coverage
go tool cover -func coverage.out
# HTML representation of tested code
go tool cover -html coverage.out
```
