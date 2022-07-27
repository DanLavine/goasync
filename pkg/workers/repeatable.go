package workers

import (
	"context"

	"github.com/DanLavine/goasync"
)

type repeatable struct {
	callback goasync.Worker
}

// Repeatable Workers are used to run any tasks that might fail with an error.
// This worker will swallow ther `Run()` error for the passed in `worker` argument
// and restart the process. Initialize and Cleanup errors are propigated as normal
func NewRepeatableWorker(worker goasync.Worker) *repeatable {
	return &repeatable{
		callback: worker,
	}
}

func (r *repeatable) Initialize() error {
	return r.callback.Initialize()
}

func (r *repeatable) Cleanup() error {
	return r.callback.Cleanup()
}

func (r *repeatable) Work(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_ = r.callback.Work(ctx)
		}
	}
}
