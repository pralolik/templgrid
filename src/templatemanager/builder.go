package templatemanager

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"html/template"
	"reflect"

	"github.com/tdewolff/minify/v2"
	htmlMinify "github.com/tdewolff/minify/v2/html"
)

const MnBlck = "email"
const SbjBlck = "subject"

func getHTMLFromTemplate(
	tmplt *template.Template,
	components []string,
	data interface{}) (res string, err error) {
	for _, component := range components {
		tmplt, err = tmplt.Parse(component)
		if err != nil {
			err = fmt.Errorf("error with component parsing %s: %w ", component, err)
			return
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = tmplt.Execute(buf, data)
	if err != nil {
		err = fmt.Errorf("error with templatemanager executing %s: %w ", tmplt.Name(), err)
		return
	}

	m := minify.New()
	m.AddFunc("text/html", htmlMinify.Minify)
	mb, err := m.Bytes("text/html", buf.Bytes())
	if err != nil {
		err = fmt.Errorf("error with minifing %s: %w ", tmplt.Name(), err)
		return
	}

	res = string(mb)
	return
}

func getDefaultFunctionsMap(i10n map[string]string) template.FuncMap {
	return template.FuncMap{
		"args": args,
		"__": func(name string, input ...interface{}) (string, error) {
			if format, ok := i10n[name]; ok {
				return fmt.Sprintf(format, input...), nil
			}
			return "", nil
		},
		"unescape": html.UnescapeString,
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

func createTemplate(txt, bn string, i10n map[string]string) (*template.Template, error) {
	t := template.New(bn)
	t.Funcs(getDefaultFunctionsMap(i10n))
	t, err := t.Parse(txt)
	if err != nil {
		return nil, fmt.Errorf("error with templatemanager parsing: %w ", err)
	}
	return t, nil
}
