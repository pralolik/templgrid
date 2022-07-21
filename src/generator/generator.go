package generator

import (
	"fmt"

	"github.com/pralolik/templgrid/src/generator/input"
	"github.com/pralolik/templgrid/src/generator/output"
	"github.com/pralolik/templgrid/src/logging"
	"github.com/pralolik/templgrid/src/resources"
)

type Generator struct {
	input   input.Interface
	outputs []output.Interface
	log     logging.Logger
}

func New(input input.Interface, output []output.Interface, log logging.Logger) *Generator {
	return &Generator{
		input:   input,
		outputs: output,
		log:     log,
	}
}

func (g *Generator) Generate() error {
	cmpnts, err := g.input.GetComponents()
	if err != nil {
		return fmt.Errorf("error with components: %w ", err)
	}

	emails, err := g.input.GetEmails()
	if err != nil {
		return fmt.Errorf("error with emails :%w ", err)
	}

	i10n, err := g.input.GetI10n()
	if err != nil {
		return fmt.Errorf("error with i10n :%w ", err)
	}
	for _, email := range emails {
		resource := &resources.TemplateResource{
			Name: email.Name,
		}
		resource.EmailTemplate = email.EmailTemplate
		for _, out := range g.outputs {
			if err = out.AddEmail(resource); err != nil {
				return fmt.Errorf("error with adding to output :%w ", err)
			}
		}
	}

	for _, out := range g.outputs {
		out.AddComponents(cmpnts)
		out.AddI10n(i10n)
		if err = out.Push(); err != nil {
			return fmt.Errorf("error with pushing to output :%w ", err)
		}
	}

	g.log.Info("Generation completed")
	return nil
}
