package flow

import (
	"bytes"
	"gograph/internal/log"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v2"
)

// FlowDefinition
// ----------------------------------------
type FlowDefinition struct {
	BasePath string

	// The name of the flow
	Name string `yaml:",omitempty"`

	// Endpoint to query for the step
	Endpoints []FlowEndpoint

	// Values extracted from the steps
	State map[string]interface{}

	// Steps to process
	Steps []FlowStep
}

func (f *FlowDefinition) String() string {
	d, err := yaml.Marshal(&f)
	if err != nil {
		log.Fatalf("FlowDefinition.String() error: %v", err)
	}
	return string(d)
}

func (f *FlowDefinition) LoadEnpoints() error {
	for _, endpoint := range f.Endpoints {
		err := endpoint.LoadSchema(f.BasePath)
		if err != nil {
			log.Println("unable to load endpoint schema")
			return err
		}
	}
	return nil
}

func NewFlowDefinition(basePath string) *FlowDefinition {
	f := &FlowDefinition{
		BasePath: basePath,
		State:    make(map[string]interface{}),
	}
	return f
}

type FlowDefImport struct {
	Import []string `yaml:"import,omitempty"`
}

func loadFlowDefinitionPart(file string) ([]byte, error) {

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// The default yaml lib doesn't seem to be able to define custom tags
	//  nor decode a file partially invalid (missing ref) so we use a dirty hack
	// Load the file a first time to be able to read the imports
	imports := FlowDefImport{}

	// Extract only the import node
	re := regexp.MustCompile(`(?m:import:\s*(\s*-.*\s*)*)`)
	importChunk := re.Find(data)

	// Parse the node
	err = yaml.Unmarshal(importChunk, &imports)
	if err != nil {
		log.Println(imports)
		return nil, err
	}
	// Read all the files and concatenate them
	var b bytes.Buffer
	writer := io.Writer(&b)
	for _, importedFile := range imports.Import {

		// Get a path relative to the file for the import
		importedFilePath, _ := filepath.Abs(file)
		importedFilePath = filepath.Dir(importedFilePath)
		if filepath.IsAbs(importedFile) {
			importedFilePath = importedFile
		} else {
			importedFilePath = filepath.Join(importedFilePath, importedFile)
		}
		imported, err := loadFlowDefinitionPart(importedFilePath)
		if err != nil {
			return nil, err
		}

		writer.Write([]byte("\n\n"))
		writer.Write(imported)
	}
	writer.Write(data)

	return b.Bytes(), nil
}

func LoadFlowDefinitionFile(file string) (*FlowDefinition, error) {
	log.Debugf("Loading flow definition file: %v", file)

	b, err := loadFlowDefinitionPart(file)
	if err != nil {
		return nil, err
	}

	basePath, _ := filepath.Abs(file)
	basePath = filepath.Dir(basePath)

	return LoadFlowDefinition(b, basePath), nil
}

func LoadFlowDefinition(data []byte, basePath string) *FlowDefinition {
	t := NewFlowDefinition(basePath)
	err := yaml.Unmarshal([]byte(data), t)
	if err != nil {
		log.Fatalln("Unable to load flow", err)
	}
	return t
}
