package tasks

import (
	"context"
	"time"

	"github.com/DanLavine/goasync"
)

type repeatTimer struct {
	repeatTime time.Duration
	subTask    goasync.Task
}

//	PARAMETERS:
//	- repeatTime - amount of time to wait between invoking the subTask
//	- task - task process that is wrapped for the periodic timer
//
// RepeatTimer Executers are used to run periodic one off tasks on a timer.
func RepeatTimer(repeatTime time.Duration, task goasync.Task) *repeatTimer {
	return &repeatTimer{
		repeatTime: repeatTime,
		subTask:    task,
	}
}

// passthough init and cleanup
func (r *repeatTimer) Initialize(ctx context.Context) error { return r.subTask.Initialize(ctx) }
func (r *repeatTimer) Cleanup(ctx context.Context) error    { return r.subTask.Cleanup(ctx) }

func (r *repeatTimer) Execute(ctx context.Context) error {
	ticker := time.NewTicker(r.repeatTime)
	defer ticker.Stop()

	// start off repeatable process by running the calback. This way we don't
	// have to wait for the timer befor making our 1st calback execution
	_ = r.subTask.Execute(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Possible race on selection, If the 'subTask.Execute(...)' takes longer than the
			// ticker, this select is random. So be sure to double check if we should really
			// shutdown
			select {
			case <-ctx.Done():
				return nil
			default:
				_ = r.subTask.Execute(ctx)
			}
		}
	}
}
