package templatemanager

import (
	"fmt"
	"strings"

	"github.com/pralolik/templgrid/src/resources"
)

type EmailStorage struct {
	templates  map[string]*resources.TemplateResource
	components []string
	i10n       map[string]map[string]string
}

func NewEmailStorage() *EmailStorage {
	return &EmailStorage{
		templates:  map[string]*resources.TemplateResource{},
		components: []string{},
	}
}

func (es *EmailStorage) AddEmail(res *resources.TemplateResource) {
	es.templates[res.Name] = res
}

func (es *EmailStorage) AddComponents(components []string) {
	es.components = components
}

func (es *EmailStorage) AddI10n(i10n map[string]map[string]string) {
	es.i10n = i10n
}

func (es *EmailStorage) HasEmail(emailName string) error {
	_, err := es.getTemplate(emailName)

	return err
}

func (es *EmailStorage) BuildEmail(emailName string, locale string, parameters interface{}) (string, string, error) {
	template, err := es.getTemplate(emailName)
	if err != nil {
		return "", "", fmt.Errorf("build email error: %w ", err)
	}

	i10n, err := es.getLocale(locale)
	if err != nil {
		return "", "", fmt.Errorf("no i10n %s found", locale)
	}

	subjectTmpl, err := createTemplate(template.EmailTemplate, SbjBlck, i10n)
	if err != nil {
		return "", "", fmt.Errorf("build email error: %w ", err)
	}

	emailTmpl, err := createTemplate(template.EmailTemplate, MnBlck, i10n)
	if err != nil {
		return "", "", fmt.Errorf("build email error: %w ", err)
	}

	subject, err := getHTMLFromTemplate(subjectTmpl, es.components, parameters)
	if err != nil {
		return "", "", fmt.Errorf("build email error: %w ", err)
	}

	email, err := getHTMLFromTemplate(emailTmpl, es.components, parameters)
	if err != nil {
		return "", "", fmt.Errorf("build email error: %w ", err)
	}

	return subject, email, nil
}

func (es *EmailStorage) getTemplate(emailName string) (*resources.TemplateResource, error) {
	if template, ok := es.templates[emailName]; ok {
		return template, nil
	}

	return nil, fmt.Errorf("no email template with name %s found", emailName)
}

func (es *EmailStorage) getLocale(locale string) (map[string]string, error) {
	if locale == "" {
		return map[string]string{}, nil
	}
	locale = strings.ToLower(locale)
	if localeParams, ok := es.i10n[locale]; ok {
		return localeParams, nil
	}

	return nil, fmt.Errorf("no locale with name %s found", locale)
}
