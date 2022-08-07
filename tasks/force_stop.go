package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/DanLavine/goasync"
)

type forceStop struct {
	duration time.Duration
	subTask  goasync.Task
}

// ForceStop Tasks are used to ensure a forceful termination even if subTasks don't properly stop executing
//
// Args:
//  - duration: time to wait for subtask to finish. If this time is reached an error is returned and subTask is ignored
//  - task: task process that is wrapped to handle interruptions
func ForceStop(duration time.Duration, task goasync.Task) *forceStop {
	return &forceStop{
		duration: duration,
		subTask:  task,
	}
}

func (fs *forceStop) Initialize() error { return fs.subTask.Initialize() }
func (fs *forceStop) Cleanup() error    { return fs.subTask.Cleanup() }

func (fs *forceStop) Execute(ctx context.Context) error {
	// make this buffered so chan can still be garbage collected if the subTask eventually exits
	errChan := make(chan error, 1)
	go func() {
		errChan <- fs.subTask.Execute(ctx)
		close(errChan)
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			select {
			case <-time.After(fs.duration):
				return fmt.Errorf("Failed to stop sub task in %v", fs.duration)
			case err := <-errChan:
				return err
			}
		}
	}
}
