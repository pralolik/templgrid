package output

import (
	"github.com/pralolik/templgrid/src/resources"
	"github.com/pralolik/templgrid/src/templatemanager"
)

type StoreOutput struct {
	storage *templatemanager.EmailStorage
}

func NewStoreOutput(storage *templatemanager.EmailStorage) *StoreOutput {
	return &StoreOutput{
		storage: storage,
	}
}

func (do *StoreOutput) AddEmail(res *resources.TemplateResource) error {
	do.storage.AddEmail(res)
	return nil
}

func (do *StoreOutput) AddComponents(components []string) {
	do.storage.AddComponents(components)
}

func (do *StoreOutput) AddI10n(i10n map[string]map[string]string) {
	do.storage.AddI10n(i10n)
}

func (do *StoreOutput) Push() error {
	return nil
}
