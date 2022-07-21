package container

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime/debug"

	"github.com/pralolik/templgrid/src/api"
	"github.com/pralolik/templgrid/src/generator"
	"github.com/pralolik/templgrid/src/generator/input"
	"github.com/pralolik/templgrid/src/generator/output"
	"github.com/pralolik/templgrid/src/logging"
	"github.com/pralolik/templgrid/src/queue"
	"github.com/pralolik/templgrid/src/sendgrid"
	"github.com/pralolik/templgrid/src/templatemanager"
)

type AppContainer struct {
	Log          logging.Logger
	Config       *Config
	EmailStorage *templatemanager.EmailStorage
}

func NewAppContainer(config *Config, log logging.Logger) (*AppContainer, error) {
	emailStorage := templatemanager.NewEmailStorage()
	gen := generator.New(getInput(log), getOutputs(config, log, emailStorage), log)
	if err := gen.Generate(); err != nil {
		return nil, fmt.Errorf("can't create app container: %w ", err)
	}

	return &AppContainer{
		Config:       config,
		Log:          log,
		EmailStorage: emailStorage,
	}, nil
}

func (cnt *AppContainer) Run(ctx context.Context) {
	defer cnt.recover(func(_ error) {
		cnt.Run(ctx)
	})
	q := cnt.createQueue()
	cnt.runQueue(ctx, q)
	cnt.runSendGrid(ctx, q)
	cnt.runAPI(ctx, q)
	<-ctx.Done()
	if ctx.Err() != nil {
		cnt.Log.Error("AppContainer shutdown with error: %v ", ctx.Err())
		return
	}
	cnt.Log.Info("AppContainer shutdown")
}

func (cnt *AppContainer) runQueue(ctx context.Context, q queue.Interface) {
	go func() {
		defer cnt.recover(func(_ error) { cnt.runQueue(ctx, q) })()
		if err := q.Run(ctx); err != nil {
			cnt.Log.Error("queue error: %v ", err)
			panic(err)
		}
	}()
}

func (cnt *AppContainer) createQueue() queue.Interface {
	return queue.NewInternalQueue(cnt.Log)
}

func (cnt *AppContainer) runSendGrid(ctx context.Context, q queue.Interface) {
	sgCfg := cnt.Config.Sendgrid
	if !sgCfg.Enabled {
		return
	}
	s := sendgrid.NewSendGrid(sgCfg.PrivateToken, sgCfg.SandBox, cnt.Log, cnt.EmailStorage)
	go func() {
		defer cnt.recover(func(_ error) { cnt.runSendGrid(ctx, q) })()
		if err := s.Run(ctx, q); err != nil {
			cnt.Log.Error("sendgrid error: %v ", err)
			panic(err)
		}
	}()
}

func (cnt *AppContainer) runAPI(ctx context.Context, q queue.Interface) {
	apiConfig := cnt.Config.APIConfig
	previewConfig := cnt.Config.APIConfig
	a := api.NewServer(
		cnt.Log,
		api.WithAPI(apiConfig.Enabled, apiConfig.APIKey, apiConfig.Port),
		api.WithPreview(previewConfig.Enabled, cnt.EmailStorage),
	)
	go func() {
		defer cnt.recover(func(err error) {
			var optE *net.OpError
			if !errors.As(err, &optE) {
				cnt.runAPI(ctx, q)
			}
			os.Exit(1)
		})()
		if err := a.Serve(ctx, q); err != nil {
			cnt.Log.Error("api error: %v ", err)
			panic(err)
		}
	}()
}

func (cnt *AppContainer) recover(f func(err error)) func() {
	return func() {
		var err error
		r := recover()
		if r != nil {
			cnt.Log.Error("AppContainer panic: %v Stack:\n%s", r, debug.Stack())
			or, ok := r.(*net.OpError)
			if ok {
				err = or
			}
		}
		f(err)
	}
}

func getOutputs(_ *Config, _ logging.Logger, emailStorage *templatemanager.EmailStorage) []output.Interface {
	var otpts []output.Interface
	otpts = append(otpts, output.NewStoreOutput(emailStorage))

	return otpts
}

func getInput(log logging.Logger) input.Interface {
	return input.NewDirectoryInput(log)
}
