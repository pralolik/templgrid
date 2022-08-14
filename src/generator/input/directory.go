package input

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pralolik/templgrid/src/helper"
	"github.com/pralolik/templgrid/src/logging"
	"github.com/pralolik/templgrid/static"
)

type DirectoryInput struct {
	components []string
	logger     logging.Logger
}

func NewDirectoryInput(logger logging.Logger) *DirectoryInput {
	return &DirectoryInput{
		logger: logger,
	}
}

func (di *DirectoryInput) GetComponents() ([]string, error) {
	if di.components != nil {
		return di.components, nil
	}
	var templates []string

	err := fs.WalkDir(static.Components(), "components", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error with scan directory :%w ", err)
		}
		if d.IsDir() {
			return nil
		}
		extension := filepath.Ext(path)
		if extension != ".html" {
			return nil
		}
		txt, err := fs.ReadFile(static.Components(), path)
		if err != nil {
			return fmt.Errorf("error with reading path %s :%w ", path, err)
		}
		templates = append(templates, string(txt))
		return nil
	})
	if err != nil {
		return nil, err
	}

	di.logger.Debug("%d components created", len(templates))

	di.components = templates
	return templates, nil
}

func (di *DirectoryInput) GetEmails() ([]*EmailInputTemplate, error) {
	var tmplts []*EmailInputTemplate
	err := fs.WalkDir(static.Emails(), "emails", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error with scan directory :%w ", err)
		}
		if d.IsDir() {
			return nil
		}
		extension := filepath.Ext(path)
		if extension != ".html" {
			return nil
		}
		tmpltName := helper.GetTemplateNameFromFile(d.Name())
		di.logger.Debug("Parsing email %s", path)
		resource := &EmailInputTemplate{
			Name: tmpltName,
		}
		txt, err := fs.ReadFile(static.Emails(), path)
		if err != nil {
			return fmt.Errorf("read file error %s: %w ", resource.Name, err)
		}
		resource.EmailTemplate = string(txt)
		resource.SubjectTemplate = string(txt)
		tmplts = append(tmplts, resource)
		return nil
	})
	if err != nil {
		return nil, err
	}

	di.logger.Debug("%d email templates created", len(tmplts))

	return tmplts, nil
}

func (di *DirectoryInput) GetI10n() (map[string]map[string]string, error) {
	i10nFiles := static.I10n()
	i10n := map[string]map[string]string{}
	err := fs.WalkDir(i10nFiles, "i10n", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error with directory scan: %w ", err)
		}
		if d.IsDir() {
			return nil
		}
		extension := filepath.Ext(path)
		if extension != ".json" {
			return nil
		}
		localeName := strings.ToLower(helper.GetTemplateNameFromFile(d.Name()))
		txt, err := fs.ReadFile(i10nFiles, path)
		if err != nil {
			return fmt.Errorf("can't read %s: %w ", localeName, err)
		}
		var i10nMap map[string]string
		err = json.Unmarshal(txt, &i10nMap)
		if err != nil {
			return fmt.Errorf("can't unmarlsahl %s: %w ", path, err)
		}
		i10n[localeName] = i10nMap
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("can't load i10n: %w ", err)
	}

	return i10n, nil
}
