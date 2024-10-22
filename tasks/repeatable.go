package tasks

import (
	"context"

	"github.com/DanLavine/goasync/v2"
)

type repeatable struct {
	callback goasync.Task
}

//	PARAMETERS:
//	- task - task process that is wrapped for the periodic timer
//
// Repeatable Tasks are used to run any tasks that might fail with an error.
// This task will swallow ther `Run()` error for the passed in `task` argument
// and restart the process. Initialize and Cleanup errors are propigated as normal
func Repeatable(task goasync.Task) *repeatable {
	return &repeatable{
		callback: task,
	}
}

// passthrough init and cleanup
func (r *repeatable) Initialize(ctx context.Context) error { return r.callback.Initialize(ctx) }
func (r *repeatable) Cleanup(ctx context.Context) error    { return r.callback.Cleanup(ctx) }

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
