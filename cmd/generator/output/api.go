package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pralolik/templgrid/cmd/generator/output/auth"
	"github.com/pralolik/templgrid/cmd/logging"
	"github.com/pralolik/templgrid/pkg"
	"io"
	"net/http"
	"strings"
)

type ApiOutput struct {
	deleteUnlisted    bool
	logger            logging.Logger
	authProvider      auth.Provider
	getPath           string
	postPath          string
	deletePath        string
	apiTemplates      map[string]*pkg.ApiTemplateResource
	preparedTemplates map[string]*pkg.ApiTemplateResource
	notListed         map[string]bool
	noUpdateNeeded    map[string]bool
}

func NewApiOutput(authProvider auth.Provider,
	deleteUnlisted bool,
	schema, host, getPath, postPath, deletePath string,
	logger logging.Logger,
) (*ApiOutput, error) {
	apiOutput := &ApiOutput{
		deleteUnlisted:    deleteUnlisted,
		logger:            logger,
		authProvider:      authProvider,
		getPath:           fmt.Sprintf("%s://%s/%s", schema, host, getPath),
		postPath:          fmt.Sprintf("%s://%s/%s", schema, host, postPath),
		deletePath:        fmt.Sprintf("%s://%s/%s", schema, host, deletePath),
		preparedTemplates: map[string]*pkg.ApiTemplateResource{},
		apiTemplates:      map[string]*pkg.ApiTemplateResource{},
		notListed:         map[string]bool{},
		noUpdateNeeded:    map[string]bool{},
	}

	if err := apiOutput.loadTemplatesFromApi(); err != nil {
		return nil, err
	}

	for _, tmpl := range apiOutput.apiTemplates {
		apiOutput.notListed[tmpl.Name] = true
	}

	return apiOutput, nil
}

func (ao *ApiOutput) Add(res *TemplateResource) error {
	name := res.Name
	isNew := ao.isNew(name)
	ao.removeFromUnlisted(name)

	if isNew {
		ao.preparedTemplates[name] = ao.createNewApiTemplateResource(res)
	} else {
		updated, needUpdate := ao.mergeWithOldApiTemplateResource(res)
		if !needUpdate {
			ao.noUpdateNeeded[name] = true
		}
		ao.preparedTemplates[name] = updated
	}

	return nil
}

func (ao *ApiOutput) Push() error {
	for _, preparedTemplate := range ao.preparedTemplates {
		if !ao.isUpdateNeeded(preparedTemplate.Name) {
			continue
		}
		if err := ao.pushNewTemplateToApi(preparedTemplate); err != nil {
			return err
		}
	}

	if ao.deleteUnlisted && len(ao.notListed) != 0 {
		for name, isNotListed := range ao.notListed {
			if isNotListed {
				if err := ao.deleteTemplateFromApi(name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (ao *ApiOutput) isNew(name string) bool {
	for tmplName := range ao.apiTemplates {
		if tmplName == name {
			return false
		}
	}

	return true
}

func (ao *ApiOutput) isUpdateNeeded(name string) bool {
	if _, ok := ao.noUpdateNeeded[name]; ok {
		return false
	}

	return true
}

func (ao *ApiOutput) removeFromUnlisted(name string) {
	delete(ao.notListed, name)
}

func (ao *ApiOutput) createNewApiTemplateResource(res *TemplateResource) *pkg.ApiTemplateResource {
	return &pkg.ApiTemplateResource{
		Name:    res.Name,
		Html:    res.Html,
		Subject: res.Subject,
	}
}

func (ao *ApiOutput) mergeWithOldApiTemplateResource(res *TemplateResource) (*pkg.ApiTemplateResource, bool) {
	existedApiResource, ok := ao.apiTemplates[res.Name]
	if !ok {
		return ao.createNewApiTemplateResource(res), true
	}
	apiResource := &pkg.ApiTemplateResource{
		InternalId:    existedApiResource.InternalId,
		ActiveVersion: existedApiResource.ActiveVersion,
		Versions:      existedApiResource.Versions,
		Name:          existedApiResource.Name,
		Html:          res.Html,
		Subject:       res.Subject,
	}
	changed := false
	if !bytes.Equal([]byte(apiResource.Html), []byte(existedApiResource.Html)) {
		changed = true
	}

	if !bytes.Equal([]byte(apiResource.Subject), []byte(existedApiResource.Subject)) {
		changed = true
	}

	return apiResource, changed

}

func (ao *ApiOutput) loadTemplatesFromApi() error {
	response, err := ao.makeRequest(http.MethodGet, ao.getPath, nil)
	if err != nil {
		return err
	}
	if response.StatusCode > 299 {
		return fmt.Errorf("impossible to load templates from api %d", response.StatusCode)
	}

	respBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("impossible to read templates from api %v", err)
	}
	var templates []*pkg.ApiTemplateResource
	err = json.Unmarshal(respBytes, &templates)
	if err != nil {
		return fmt.Errorf("impossible to read templates from api %v", err)
	}
	for _, tmpl := range templates {
		ao.apiTemplates[tmpl.Name] = tmpl
	}

	return nil
}

func (ao *ApiOutput) pushNewTemplateToApi(new *pkg.ApiTemplateResource) error {
	body, err := json.Marshal(new)
	if err != nil {
		return fmt.Errorf("impossible to marhsal new template %s to api %v", new.Name, err)
	}
	response, err := ao.makeRequest(http.MethodPost, ao.postPath, body)
	if err != nil {
		return err
	}

	if response.StatusCode > 299 {
		return fmt.Errorf("impossible to push new template %s to api %d", new.Name, response.StatusCode)
	}

	return nil
}

func (ao *ApiOutput) deleteTemplateFromApi(toDeleteSlug string) error {
	response, err := ao.makeRequest(http.MethodDelete, strings.ReplaceAll(ao.deletePath, "{slug}", toDeleteSlug), nil)
	if err != nil {
		return err
	}

	if response.StatusCode > 299 {
		return fmt.Errorf("impossible to delete template %s from api %s", toDeleteSlug, response.StatusCode)
	}

	return nil
}

func (ao *ApiOutput) makeRequest(method string, url string, body []byte) (*http.Response, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
