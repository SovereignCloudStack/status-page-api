# Status Page API

The server handles components, incidents and other resources according to the [Open API spec for status page](https://github.com/SovereignCloudStack/status-page-openapi).

## Config

The service can be configured by  either providing flags to the binary or via environment variables.

### Examples

Changeing the listining address to `localhost:3001`.

#### Flags

```bash
./bin/status-page-api --listen-address :3001
```

#### Environment

Every config key is prefixed by `STATUS_PAGE`.

```bash
STATUS_PAGE_LISTEN_ADDRESS=:3001 ./bin/status-page-api
```

### Verbosity

Default logging level is set to `WARN` level and can be increased by adding `-v` to the binary, to a maximum of `-vvv` being the `TRACE` level.

Setting verbosity by environment variables requires setting it to a number.

```env
# INFO level
STATUS_PAGE_VERBOSE=1
# DEBUG level
STATUS_PAGE_VERBOSE=2
# TRACE level
STATUS_PAGE_VERBOSE=3
```

## Development settings

Create `secrets.env` for config:

```env
STATUS_PAGE_VERBOSE=3
STATUS_PAGE_SWAGGER_UI_ENABLED=true
STATUS_PAGE_LISTEN_ADDRESS=:3000
STATUS_PAGE_CORS_ORIGINS="localhost 127.0.0.1"
STATUS_PAGE_DATABASE_CONNECTION_STRING="host=127.0.0.1 user=postgres dbname=postgres port=5432 password=debug sslmode=disable"
STATUS_PAGE_PROVISIONING_FILE=./provisioning.yaml

```

`STATUS_PAGE_SWAGGER_UI_ENABLED` enables the [Swagger Web UI](https://swagger.io/tools/swagger-ui/) for local debugging purposes, which is disabled by default.

### Note

Source the env before executing the binary, to configure the service.

The `Makefile` target `make serve` sources the env file `secrets.env` by default.
