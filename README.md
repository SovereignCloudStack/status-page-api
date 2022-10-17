# API Server Component of SCS Status Page

Implemented in Go, it can be built with usual Go tooling.

## Local debugging setup

```bash
# 1. Start up Postgres compatible database
podman run --network host -e POSTGRES_PASSWORD=debug -d postgres
# 2. Run API, read "help" for parameters
go run *.go --help
```
