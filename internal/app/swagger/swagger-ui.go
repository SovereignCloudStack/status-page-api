package swagger

import (
	"net/http"

	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/labstack/echo/v4"
)

// https://github.com/swagger-api/swagger-ui/blob/master/docs/usage/installation.md
const swaggerHTML = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta
      name="description"
      content="SwaggerUI"
    />
    <title>SwaggerUI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
  </head>
  <body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js" crossorigin></script>
  <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-standalone-preset.js" crossorigin></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        validatorUrl: null,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        layout: "StandaloneLayout",
      });
    };
  </script>
  </body>
</html>
`

// ServeSwagger serves the html for displaying swagger UI.
func ServeSwagger(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, swaggerHTML) //nolint:wrapcheck
}

// ServeOpenAPISpec decodes and serves the OpenAPI.json spec.
func ServeOpenAPISpec(ctx echo.Context) error {
	swagger, err := apiServerDefinition.GetSwagger()
	if err != nil {
		ctx.Logger().Error(err)

		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSONPretty(http.StatusOK, swagger, "  ") //nolint:wrapcheck
}
