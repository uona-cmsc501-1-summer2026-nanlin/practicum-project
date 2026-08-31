package swagger

import (
	"fmt"
	"io"
	"os"

	"github.com/pb33f/libopenapi"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Spec holds the parsed OpenAPI document for rendering at runtime.
type Spec struct {
	data    []byte
	v3model *v3high.Document
}

// Load reads and parses the merged OpenAPI file.
func Load(specPath string) (*Spec, error) {
	file, err := os.Open(specPath)
	if err != nil {
		return nil, fmt.Errorf("open spec: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read spec: %w", err)
	}

	document, docErrs := libopenapi.NewDocument(data)
	if docErrs != nil {
		return nil, fmt.Errorf("parse spec: %w", docErrs)
	}

	v3model, err := document.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("build v3 model: %w", err)
	}

	return &Spec{
		data:    data,
		v3model: &v3model.Model,
	}, nil
}

// WithUpdate applies runtime mutations (title, version, servers, description).
func (s *Spec) WithUpdate(fn func(m *v3high.Document)) *Spec {
	if fn != nil && s.v3model != nil {
		fn(s.v3model)
	}
	return s
}

// Render returns the current spec as YAML/JSON bytes.
func (s *Spec) Render() []byte {
	if s.v3model != nil {
		out, _ := s.v3model.Render()
		return out
	}
	return s.data
}
