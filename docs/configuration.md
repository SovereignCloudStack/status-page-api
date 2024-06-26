# Configuration

Configuration can be done by environment variables or flags to the binary.

Code to the configuration can be found at `internal/app/config/config.go`.

| Environment key                        | Flag                         | Description                                  | Type         | Default              |
| -------------------------------------- | ---------------------------- | -------------------------------------------- | ------------ | -------------------- |
| **General settings**                   |                              |                                              |              |                      |
| STATUS_PAGE_PROVISIONING-FILE          | --provisioning-file          | YAML file containing the initial values      | Path         | ./provisioning.yaml  |
| STATUS_PAGE_SHUTDOWN_TIMEOUT           | --shutdown-timeout           | Timeout to gracefully stop the server        | Duration     | 10s                  |
| STATUS_PAGE_VERBOSE                    | -v / --verbose               | Increase log level                           | Counter      | 0                    |
| **Server settings**                    |                              |                                              |              |                      |
| STATUS_PAGE_SERVER_ADDRESS             | --server-address             | API server listen address                    | String       | :3000                |
| STATUS_PAGE_SERVER_ALLOWED_ORIGINS     | --server-allowed-origins     | List of allowed CORS origins                 | String Array | 127.0.0.1, localhost |
| STATUS_PAGE_SERVER_SWAGGER_UI_ENABLED  | --server-swagger-ui-enabled  | Enable the swagger UI at `/swagger`          | Boolean      | False                |
| **Database settings**                  |                              |                                              |              |                      |
| STATUS_PAGE_DATABASE_CONNECTION_STRING | --database-connection-string | PostgreSQL connection string                 | String       |                      |
| **Metrics settings**                   |                              |                                              |              |                      |
| STATUS_PAGE_METRICS_ADDRESS            | --metrics-address            | Enable and set metrics server listen address | String       |                      |
| STATUS_PAGE_METRICS_NAMESPACE          | --metrics-namespace          | Metrics namespace                            | String       | status_page          |
| STATUS_PAGE_METRICS_SUBSYSTEM          | --metrics-subsystem          | Metrics subsystem name                       | String       | api                  |
