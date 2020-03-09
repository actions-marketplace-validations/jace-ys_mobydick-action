package workflow

import (
	"fmt"
	"io/ioutil"
)

type WorkflowFile struct {
	Path    string
	Content []byte
}

func NewWorkflowFile(file string) (*WorkflowFile, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return &WorkflowFile{
		Path:    fmt.Sprintf(".github/workflows/%s", file),
		Content: content,
	}, nil
}
