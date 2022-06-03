package pkg

type ApiTemplateResource struct {
	Name          string   `json:"name"`
	InternalId    string   `json:"internal_id"`
	Subject       string   `json:"subject"`
	Html          string   `json:"html"`
	Parameters    []string `json:"parameters"`
	Versions      []string `json:"versions"`
	ActiveVersion string   `json:"active_version"`
}
