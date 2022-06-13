package input

import (
	"fmt"
	"github.com/pralolik/templgrid/cmd/logging"
	"io/fs"
	"path/filepath"
)

type DirectoryInput struct {
	components     []string
	emailPath      string
	componentsPath string
	prefix         string
	logger         logging.Logger
}

func NewDirectoryInput(prefix, emailPath, componentsPath string, logger logging.Logger) *DirectoryInput {
	return &DirectoryInput{
		prefix:         prefix,
		emailPath:      emailPath,
		componentsPath: componentsPath,
		logger:         logger,
	}
}

func (di *DirectoryInput) GetComponents() ([]string, error) {
	if di.components != nil {
		return di.components, nil
	}
	var includeFiles []string
	err := filepath.Walk(di.componentsPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error with scan of components directory %v", err)
		}
		if info.IsDir() {
			return nil
		}
		includeFiles = append(includeFiles, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	var templates []string
	for _, file := range includeFiles {
		di.logger.Debug("Parsing component template %s", file)
		txt, err := getFileData(file)
		if err != nil {
			return nil, err
		}
		templates = append(templates, txt)
	}

	di.logger.Debug("%d components created", len(templates))

	di.components = templates
	return templates, nil
}

func (di *DirectoryInput) GetEmails() ([]*EmailInputTemplate, error) {
	var includeFiles []string
	err := filepath.Walk(di.emailPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error with scan of email directory %v", err)
		}
		if info.IsDir() {
			return nil
		}
		includeFiles = append(includeFiles, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var tmplts []*EmailInputTemplate
	for _, file := range includeFiles {
		tmpltName := fmt.Sprintf("%s%s", di.prefix, getTemplateNameFromFile(file))
		di.logger.Debug("Parsing email template %s", file)
		resource := &EmailInputTemplate{
			Name: tmpltName,
		}
		txt, err := getFileData(file)
		if err != nil {
			return nil, fmt.Errorf("%s %v", resource.Name, err)
		}

		if resource.EmailTemplate, err = createTemplate(txt, MnBlck); err != nil {
			return nil, fmt.Errorf("%s %v", resource.Name, err)
		}

		if resource.SubjectTemplate, err = createTemplate(txt, SbjBlck); err != nil {
			return nil, fmt.Errorf("%s %v", resource.Name, err)
		}
		tmplts = append(tmplts, resource)
	}

	di.logger.Debug("%d email templates created", len(tmplts))

	return tmplts, nil
}
