package queue

import (
	"context"

	"github.com/pralolik/templgrid/pkg"
)

type Interface interface {
	Push(entity *pkg.TemplgridEmailEntity) error
	GetChannel() (chan *pkg.TemplgridEmailEntity, error)
	Run(ctx context.Context) error
}
