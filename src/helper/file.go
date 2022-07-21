package helper

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
)

func GetTemplateNameFromFile(file string) string {
	fileName := filepath.Base(file)

	return strcase.ToCamel(strings.TrimSuffix(fileName, path.Ext(fileName)))
}
