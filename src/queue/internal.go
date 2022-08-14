package queue

import (
	"context"

	"github.com/pralolik/templgrid/pkg"
	"github.com/pralolik/templgrid/src/logging"
)

type InternalQueue struct {
	log          logging.Logger
	queueChannel chan *pkg.TemplgridEmailEntity
}

func NewInternalQueue(log logging.Logger) *InternalQueue {
	return &InternalQueue{
		log:          log,
		queueChannel: make(chan *pkg.TemplgridEmailEntity),
	}
}

func (q *InternalQueue) Push(entity *pkg.TemplgridEmailEntity) error {
	go func() { q.queueChannel <- entity }()
	return nil
}

func (q *InternalQueue) GetChannel() (chan *pkg.TemplgridEmailEntity, error) {
	return q.queueChannel, nil
}

func (q *InternalQueue) Run(ctx context.Context) error {
	<-ctx.Done()
	close(q.queueChannel)
	return nil
}
