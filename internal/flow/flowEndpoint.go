package flow

import (
	"gograph/internal/schema"
	"gograph/internal/template"
	"path/filepath"
)

// FlowEndPoint
// ----------------------------------------
type FlowEndpoint struct {
	Name       string
	SchemaFile string `yaml:"schema,omitempty"`
	Url        string `yaml:",omitempty"`

	schema *schema.Schema
}

func (e *FlowEndpoint) LoadSchema(basePath string) error {
	if e.schema == nil {

		target := e.SchemaFile

		if !filepath.IsAbs(e.SchemaFile) {
			target = filepath.Join(basePath, target)
		}

		schema, err := schema.LoadSchemaFromGlob(target)
		if err != nil {
			return err
		}
		e.schema = schema
	}
	return nil
}

func (e *FlowEndpoint) Schema() *schema.Schema {
	return e.schema
}

func (e *FlowEndpoint) UrlParsed(context *StepTemplateContext) string {
	return template.RunTemplateOrUnparsed(e.Url, context)
}
