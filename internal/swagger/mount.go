package swagger

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/gofiber/fiber/v2"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const uiPath = "swagger"

// Config controls how Swagger UI is mounted.
type Config struct {
	SpecFile     string // merged openapi.yaml
	MarkdownFile string // injected into info.description
	BaseURL      string // e.g. http://localhost:55555
	Title        string
	Version      string
}

// DefaultConfig returns practicum-project defaults.
func DefaultConfig() Config {
	return Config{
		SpecFile:     "docs/swagger/generate/openapi.yaml",
		MarkdownFile: "docs/swagger/markdown/app_description.md",
		BaseURL:      "http://localhost:55555",
		Title:        "Split It API",
		Version:      "0.2.0",
	}
}

// Mount registers GET /swagger and GET /swagger/specification on the Fiber app.
func Mount(app *fiber.App, cfg Config) error {
	spec, err := Load(cfg.SpecFile)
	if err != nil {
		return err
	}

	desc := ""
	if b, readErr := os.ReadFile(cfg.MarkdownFile); readErr == nil {
		desc = string(b)
	}

	spec.WithUpdate(func(m *v3high.Document) {
		if m.Info == nil {
			m.Info = &base.Info{}
		}
		m.Info.Title = cfg.Title
		m.Info.Version = cfg.Version
		if desc != "" {
			m.Info.Description = desc
		}
		m.Servers = []*v3high.Server{{URL: cfg.BaseURL}}
	})

	specURL := cfg.BaseURL + "/" + path.Join(uiPath, "specification")

	app.Get("/"+uiPath+"/specification", func(c *fiber.Ctx) error {
		return c.Type("json").Send(spec.Render())
	})

	tpl := template.Must(template.New("swagger").Parse(uiTemplate))
	app.Get("/"+uiPath, func(c *fiber.Ctx) error {
		c.Type("html")
		return tpl.Execute(c.Response().BodyWriter(), map[string]string{
			"Title":     cfg.Title,
			"UIVersion": uiVersion,
			"SpecURL":   specURL,
		})
	})

	app.Get("/"+uiPath+"*", func(c *fiber.Ctx) error {
		return c.Redirect("/"+uiPath, fiber.StatusPermanentRedirect)
	})

	return nil
}

// MustMount calls Mount and panics on error (for main).
func MustMount(app *fiber.App, cfg Config) {
	if err := Mount(app, cfg); err != nil {
		panic(fmt.Sprintf("swagger: %v", err))
	}
}
