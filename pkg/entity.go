package pkg

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const MaxPersonalizationPerRequest = 1000

var (
	ErrIncorrectTemplateName        = fmt.Errorf("incorrect value for %s", "template_name")
	ErrIncorrectPersonalization     = fmt.Errorf("incorrect value for %s", "send_grid_parameters.personalization")
	ErrIncorrectPersonalizationFrom = fmt.Errorf("incorrect value for %s", "send_grid_parameters.personalization.*.from")
	ErrIncorrectPersonalizationTo   = fmt.Errorf("incorrect value for %s", "send_grid_parameters.personalization.*.to")
)

type TemplgridEmailEntity struct {
	TemplateName       string        `json:"template_name"`
	Locale             string        `json:"locale,omitempty"`
	EmailParameters    interface{}   `json:"email_parameters"`
	SendGridParameters mail.SGMailV3 `json:"send_grid_parameters"`
}

func (t *TemplgridEmailEntity) Validate() error {
	if t.TemplateName == "" {
		return ErrIncorrectTemplateName
	}

	hasFrom := false
	if t.SendGridParameters.From != nil && t.SendGridParameters.From.Address != "" {
		hasFrom = true
	}

	if t.SendGridParameters.Personalizations == nil ||
		len(t.SendGridParameters.Personalizations) == 0 ||
		len(t.SendGridParameters.Personalizations) > MaxPersonalizationPerRequest {
		return ErrIncorrectPersonalization
	}

	for _, ps := range t.SendGridParameters.Personalizations {
		for _, psT := range ps.To {
			if psT.Address == "" {
				return ErrIncorrectPersonalizationTo
			}
		}

		if !hasFrom && ps.From != nil && ps.From.Address == "" {
			return ErrIncorrectPersonalizationFrom
		}
	}

	return nil
}
