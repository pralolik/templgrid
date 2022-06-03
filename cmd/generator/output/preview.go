package output

import (
	"fmt"
	"github.com/pralolik/templgrid/cmd/logging"
	"io/ioutil"
)

const PreviewFileName = "preview.html"

type PreviewOutput struct {
	path   string
	logger logging.Logger
	result string
}

func NewPreviewOutput(path string, logger logging.Logger) *PreviewOutput {
	return &PreviewOutput{
		path:   path,
		logger: logger,
	}
}

func (do *PreviewOutput) Add(res *TemplateResource) error {
	do.result += fmt.Sprintf("Name: %s Subject: %s %s", res.Name, res.Subject, res.Html)
	return nil
}

func (do *PreviewOutput) Push() error {
	fileName := do.getPath()
	err := ioutil.WriteFile(fileName, []byte(do.result), 0644)
	if err != nil {
		return err
	}

	do.logger.Info("Preview pushed to file %s", fileName)
	return nil
}

func (do *PreviewOutput) getPath() string {
	return fmt.Sprintf("%s/%s", do.path, PreviewFileName)
}
