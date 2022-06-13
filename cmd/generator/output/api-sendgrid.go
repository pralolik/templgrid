package output

import (
	"encoding/json"
	"fmt"
	"github.com/pralolik/templgrid/pkg"
	"github.com/sendgrid/sendgrid-go"
	"net/http"
	"time"
)

type SendGridOutput struct {
	api                *ApiOutput
	host               string
	apiKey             string
	activateNewVersion bool
}

func NewSendGridOutput(apiOutput *ApiOutput, host, apiKey string, activateNewVersion bool) *SendGridOutput {
	return &SendGridOutput{
		api:                apiOutput,
		apiKey:             apiKey,
		activateNewVersion: activateNewVersion,
		host:               host,
	}
}

func (sgo *SendGridOutput) Add(res *TemplateResource) error {
	return sgo.api.Add(res)
}

func (sgo *SendGridOutput) Push() error {
	for _, preparedTemplate := range sgo.api.preparedTemplates {
		if !sgo.api.isUpdateNeeded(preparedTemplate.Name) {
			continue
		}
		if sgo.api.isNew(preparedTemplate.Name) {
			err := sgo.createNewTemplateWithNewVersion(preparedTemplate)
			if err != nil {
				return err
			}
		} else {
			err := sgo.createNewTemplateVersion(preparedTemplate)
			if err != nil {
				return err
			}
		}
		if err := sgo.api.pushNewTemplateToApi(preparedTemplate); err != nil {
			return err
		}
	}

	if sgo.api.deleteUnlisted && len(sgo.api.notListed) != 0 {
		for name, isNotListed := range sgo.api.notListed {
			toDelete, ok := sgo.api.apiTemplates[name]
			if isNotListed && ok {
				if toDelete.InternalId != "" {
					if err := sgo.deleteTemplateFromSendGrid(toDelete.InternalId); err != nil {
						return err
					}
				}
				if err := sgo.api.deleteTemplateFromApi(name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type templateCreationRequest struct {
	Name       string `json:"name"`
	Generation string `json:"generation"`
}

type templateCreationResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Generation string `json:"generation"`
}

func (sgo *SendGridOutput) createNewTemplateWithNewVersion(tmpl *pkg.ApiTemplateResource) error {
	request := sendgrid.GetRequest(sgo.apiKey, "/v3/templates", sgo.host)
	request.Method = http.MethodPost
	reqBody := &templateCreationRequest{
		Name:       tmpl.Name,
		Generation: "dynamic",
	}
	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	request.Body = reqJson
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	var resp templateCreationResponse
	err = json.Unmarshal([]byte(response.Body), &resp)
	if err != nil {
		return err
	}
	tmpl.InternalId = resp.Id
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("")
	}

	err = sgo.createNewTemplateVersion(tmpl)
	if err != nil {
		return err
	}

	return nil
}

type versionRequest struct {
	TemplateId           string `json:"template_id"`
	Active               int8   `json:"active"`
	Name                 string `json:"name"`
	HtmlContent          string `json:"html_content"`
	Subject              string `json:"subject"`
	Editor               string `json:"editor"`
	GeneratePlainContent bool   `json:"generate_plain_content"`
	PlainContent         string `json:"plain_content,omitempty"`
}

func (sgo *SendGridOutput) createNewTemplateVersion(tmpl *pkg.ApiTemplateResource) error {
	if tmpl.InternalId == "" {
		return fmt.Errorf("")
	}
	request := sendgrid.GetRequest(sgo.apiKey, fmt.Sprintf("/v3/templates/%s/versions", tmpl.InternalId), sgo.host)
	request.Method = http.MethodPost
	var active int8 = 0
	if sgo.activateNewVersion == true {
		active = 1
	}
	reqBody := &versionRequest{
		TemplateId:           tmpl.InternalId,
		Active:               active,
		Name:                 time.Now().UTC().String(),
		Subject:              tmpl.Subject,
		HtmlContent:          tmpl.Html,
		Editor:               "code",
		GeneratePlainContent: false,
		PlainContent:         "",
	}
	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	request.Body = reqJson
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("")
	}

	return nil
}

func (sgo *SendGridOutput) deleteTemplateFromSendGrid(internalId string) error {
	request := sendgrid.GetRequest(sgo.apiKey, fmt.Sprintf("/v3/templates/%s", internalId), sgo.host)
	request.Method = http.MethodDelete
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("")
	}

	return nil
}
