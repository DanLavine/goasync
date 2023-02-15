package tasks

import (
	"context"
)

type running struct {
	started chan struct{}
}

// Running can be used to to if a TaskManager has started executing background threads. This
// is useful to know is a task manager received a HUP operation and all tasks were restared.
// passing in a new Running task will let us know when the task manager has started again
//
// NOTE: that this subtask will exit so its intended to be used with a "Relaxed" task manager.
//
// Args:
// * started - chan that will be closed when Execute(...) is called
func Running(started chan struct{}) *running {
	return &running{
		started: started,
	}
}

func (r *running) Initialize() error { return nil }
func (r *running) Cleanup() error    { return nil }
func (r *running) Execute(ctx context.Context) error {
	close(r.started)

	return nil
}
