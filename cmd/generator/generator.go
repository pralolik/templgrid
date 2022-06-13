package generator

import (
	"bytes"
	"fmt"
	"github.com/pralolik/templgrid/cmd/generator/input"
	"github.com/pralolik/templgrid/cmd/generator/output"
	"github.com/pralolik/templgrid/cmd/logging"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"html/template"
)

type Generator struct {
	input   input.Interface
	outputs []output.Interface
	logger  logging.Logger
}

func New(input input.Interface, outputs []output.Interface, logger logging.Logger) *Generator {
	return &Generator{
		input:   input,
		outputs: outputs,
		logger:  logger,
	}
}

func (g *Generator) Generate() error {
	cmpnts, err := g.input.GetComponents()
	if err != nil {
		return fmt.Errorf("error with components %v", err)
	}

	emails, err := g.input.GetEmails()
	if err != nil {
		return fmt.Errorf("error with emails %v", err)
	}

	for _, email := range emails {
		resource := &output.TemplateResource{
			Name: email.Name,
		}

		if resource.Html, err = getHtmlFromTemplate(email.EmailTemplate, cmpnts); err != nil {
			return fmt.Errorf("error with email template %s %v", email.Name, err)
		}

		if resource.Subject, err = getHtmlFromTemplate(email.SubjectTemplate, cmpnts); err != nil {
			return fmt.Errorf("error with subject template %s %v", email.Name, err)
		}
		g.logger.Info("Adding template %s", resource.Name)
		g.logger.Debug("Resulted resource %s", resource)

		for _, out := range g.outputs {
			if err = out.Add(resource); err != nil {
				return fmt.Errorf("error with adding to output %v", err)
			}
		}
	}

	for _, out := range g.outputs {
		if err = out.Push(); err != nil {
			return fmt.Errorf("error with pushing to output %v", err)
		}
	}

	g.logger.Info("Generation completed")
	return nil
}

func getHtmlFromTemplate(tmplt *template.Template, components []string) (res string, err error) {
	for _, component := range components {
		tmplt, err = tmplt.Parse(component)
		if err != nil {
			err = fmt.Errorf("error with component parsing %s %v", component, err)
			return
		}
	}
	buf := bytes.NewBuffer([]byte{})
	err = tmplt.Execute(buf, nil)
	if err != nil {
		err = fmt.Errorf("error with template executing %s %v", tmplt.Name(), err)
		return
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	mb, err := m.Bytes("text/html", buf.Bytes())
	if err != nil {
		err = fmt.Errorf("error with minifing %s %v", tmplt.Name(), err)
		return
	}

	res = string(mb)
	return
}
