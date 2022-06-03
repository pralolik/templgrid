package input

import (
	"errors"
	"fmt"
	"github.com/iancoleman/strcase"
	"html/template"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

const MnBlck = "email"
const SbjBlck = "subject"

type Interface interface {
	GetComponents() ([]string, error)
	GetEmails() ([]*EmailInputTemplate, error)
}

type EmailInputTemplate struct {
	Name            string
	EmailTemplate   *template.Template
	SubjectTemplate *template.Template
	Parameters      []string
}

func getDefaultFunctionsMap(inputTemplate *EmailInputTemplate) template.FuncMap {
	return template.FuncMap{
		"sendGridParam": func(name string) string {
			inputTemplate.Parameters = append(inputTemplate.Parameters, name)
			return fmt.Sprintf("{{ %s }}", name)
		},
		"args": args,
	}
}

func args(keyValues ...interface{}) (map[string]interface{}, error) {
	length := len(keyValues)
	if length%2 != 0 {
		return nil, errors.New("function args requires even parameters count")
	}
	var paramsMap = map[string]interface{}{}
	for i := 0; i < length; i += 2 {
		if reflect.TypeOf(keyValues[i]).String() != "string" {
			return nil, errors.New("function args requires key as string")
		}
		paramsMap[keyValues[i].(string)] = keyValues[i+1]
	}

	return paramsMap, nil
}

func createTemplate(r *EmailInputTemplate, txt, bn string) (*template.Template, error) {
	t := template.New(bn)
	t.Funcs(getDefaultFunctionsMap(r))
	t, err := t.Parse(txt)
	if err != nil {
		return nil, fmt.Errorf("error with template parsing %v", err)
	}
	return t, nil
}

func getFileData(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error with file reading %s %v", filePath, err)
	}

	return string(b), err
}

func getTemplateNameFromFile(file string) string {
	fileName := filepath.Base(file)

	return strcase.ToCamel(strings.TrimSuffix(fileName, path.Ext(fileName)))
}
