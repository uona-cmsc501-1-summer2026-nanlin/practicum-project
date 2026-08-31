package swagger

const uiVersion = "5.29.0"

const uiTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/{{.UIVersion}}/swagger-ui.min.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/{{.UIVersion}}/swagger-ui-bundle.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/{{.UIVersion}}/swagger-ui-standalone-preset.min.js"></script>
    <script>
      window.onload = function () {
        SwaggerUIBundle({
          url: "{{.SpecURL}}",
          dom_id: "#swagger-ui",
          presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
          plugins: [SwaggerUIBundle.plugins.DownloadUrl],
          layout: "StandaloneLayout"
        });
      };
    </script>
  </body>
</html>`
