package workers

import (
	"context"
	"time"

	"github.com/DanLavine/goasync"
)

type repeatTimer struct {
	repeatTime time.Duration
	callback   goasync.Worker
}

// RepeatTimer Workers are used to run periodic one off tasks on a timer.
// Args:
//  - repeatTime: amount of time to wait between invoking the callback
//  - callback: callback worker process that is wrapped for the periodic timer
func RepeatTimer(repeatTime time.Duration, callback goasync.Worker) *repeatTimer {
	return &repeatTimer{
		repeatTime: repeatTime,
		callback:   callback,
	}
}

func (r *repeatTimer) Initialize() error {
	return r.callback.Initialize()
}

func (r *repeatTimer) Cleanup() error {
	return r.callback.Cleanup()
}

func (r *repeatTimer) Work(ctx context.Context) error {
	ticker := time.NewTicker(r.repeatTime)
	defer ticker.Stop()

	// start off repeatable process by running the calback. This way we don't
	// have to wait for the timer befor making our 1st calback execution
	_ = r.callback.Work(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Possible race on selection, If the 'callback.Work(...)' takes longer than the
			// ticker, this select is random. So be sure to double check if we should really
			// shutdown
			select {
			case <-ctx.Done():
				return nil
			default:
				_ = r.callback.Work(ctx)
			}
		}
	}
}
