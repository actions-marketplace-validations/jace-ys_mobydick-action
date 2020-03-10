package action

import (
	"bytes"
	"fmt"
	"text/template"
)

type WorkflowFile struct {
	Path    string
	Content []byte
}

func NewWorkflowFile(file, version string) (*WorkflowFile, error) {
	tmpl, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}

	variables := struct {
		Version string
	}{
		Version: version,
	}

	var content bytes.Buffer
	err = tmpl.Execute(&content, variables)
	if err != nil {
		return nil, err
	}

	return &WorkflowFile{
		Path:    fmt.Sprintf(".github/workflows/%s", file),
		Content: content.Bytes(),
	}, nil
}
