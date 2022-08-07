package tasks

import (
	"context"

	"github.com/DanLavine/goasync"
)

type repeatable struct {
	callback goasync.Task
}

// Repeatable Tasks are used to run any tasks that might fail with an error.
// This task will swallow ther `Run()` error for the passed in `task` argument
// and restart the process. Initialize and Cleanup errors are propigated as normal
func Repeatable(task goasync.Task) *repeatable {
	return &repeatable{
		callback: task,
	}
}

// passthrough init and cleanup
func (r *repeatable) Initialize() error { return r.callback.Initialize() }
func (r *repeatable) Cleanup() error    { return r.callback.Cleanup() }

func (r *repeatable) Execute(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_ = r.callback.Execute(ctx)
		}
	}
}
