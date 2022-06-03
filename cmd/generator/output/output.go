package output

type TemplateResource struct {
	Name       string   `json:"name"`
	Subject    string   `json:"subject"`
	Html       string   `json:"html"`
	Parameters []string `json:"parameters"`
}

type Interface interface {
	Add(res *TemplateResource) error
	Push() error
}
