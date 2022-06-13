package output

type TemplateResource struct {
	Name    string `json:"name"`
	Subject string `json:"subject"`
	Html    string `json:"html"`
}

type Parameter struct {
	Optional   bool
	Name       string
	AccessPath string
}

type Interface interface {
	Add(res *TemplateResource) error
	Push() error
}
