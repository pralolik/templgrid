package output

import "github.com/pralolik/templgrid/src/resources"

type Interface interface {
	AddEmail(res *resources.TemplateResource) error
	AddComponents([]string)
	AddI10n(map[string]map[string]string)
	Push() error
}
