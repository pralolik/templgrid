package input

type Interface interface {
	GetComponents() ([]string, error)
	GetEmails() ([]*EmailInputTemplate, error)
	GetI10n() (map[string]map[string]string, error)
}

type EmailInputTemplate struct {
	Name            string
	EmailTemplate   string
	SubjectTemplate string
}
