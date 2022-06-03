package output

import (
	"fmt"
	"github.com/pralolik/templgrid/cmd/logging"
	"io/ioutil"
)

type DirectoryOutput struct {
	path   string
	logger logging.Logger
	tmplts []*TemplateResource
}

func NewDirectoryOutput(path string, logger logging.Logger) *DirectoryOutput {
	return &DirectoryOutput{
		path:   path,
		logger: logger,
	}
}

func (do *DirectoryOutput) Add(res *TemplateResource) error {
	do.tmplts = append(do.tmplts, res)
	return nil
}

func (do *DirectoryOutput) Push() error {
	for _, tplt := range do.tmplts {
		fileName := do.getPath(tplt.Name)
		err := ioutil.WriteFile(fileName, []byte(tplt.Html), 0644)
		if err != nil {
			return err
		}
		do.logger.Info("Template %s pushed to file %s", tplt.Name, fileName)
	}
	return nil
}

func (do *DirectoryOutput) getPath(tmpltName string) string {
	return fmt.Sprintf("%s/%s.%s", do.path, tmpltName, "html")
}
