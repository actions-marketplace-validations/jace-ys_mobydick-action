package workflow

import (
	"fmt"
	"io/ioutil"
)

type WorkflowFile struct {
	Content []byte
	Path    string
}

func NewWorkflowFile(file string) (*WorkflowFile, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return &WorkflowFile{
		Content: content,
		Path:    fmt.Sprintf(".github/workflows/%s", file),
	}, nil
}
