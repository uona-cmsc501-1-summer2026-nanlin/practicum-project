.PHONY: oas test run

OAS_SRC := docs/swagger/oas.yaml
OAS_OUT := docs/swagger/generate/openapi.yaml

# Redocly bundles split YAML into the file served at /swagger/specification.
oas:
	npx --yes @redocly/cli bundle $(OAS_SRC) -o $(OAS_OUT)

test:
	go test ./...

run:
	go run .
