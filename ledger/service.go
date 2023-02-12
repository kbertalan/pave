package ledger

import (
	"context"
	"fmt"

	_ "github.com/tigerbeetledb/tigerbeetle-go"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"encore.app/ledger/tb"
	"encore.app/ledger/workflow/transfer"
)

const (
	paveTaskQueue        = "pave-tq"
	transferWorkflowName = "transferWorkflow"
)

//encore:service
type Service struct {
	client     client.Client
	worker     worker.Worker
	activities *tb.Service
}

func initService() (*Service, error) {
	tbFactory := cfg.TigerBeetle.NewFactory()
	err := tbFactory.RegisterDemoAccounts(10)
	if err != nil {
		return nil, fmt.Errorf("creating demo accounts failed: %v", err)
	}

	tbActivities := tb.NewTigerBeetleActivities(tbFactory)

	c, err := client.Dial(client.Options{
		HostPort: cfg.Temporal.HostPort(),
	})
	if err != nil {
		return nil, fmt.Errorf("create temporal client failed: %v", err)
	}

	w := worker.New(c, paveTaskQueue, worker.Options{})
	w.RegisterWorkflow(transfer.Workflow)
	w.RegisterWorkflow(transfer.CancelWorkflow)
	w.RegisterActivity(tbActivities)

	err = w.Start()
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker failed: %v", err)
	}

	return &Service{client: c, worker: w, activities: tbActivities}, nil
}

func (s *Service) Shutdown(force context.Context) {
	s.client.Close()
	s.worker.Stop()
}
