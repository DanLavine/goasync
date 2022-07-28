package workers

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/DanLavine/goasync"
)

type interruptible struct {
	signals  []os.Signal
	callback goasync.Worker
}

// Interruptible Workers are used to run any tasks that can be reloaded at runtime.
// When this worker recieves an interrupt signal it will cancel the child Worker process
// and call Cleanup. Then this will call Initialize and Work again for the child Worker.
//
// Args:
//  - signals: OS signals to capture and initiate the reload operation
//  - worker: worker process that is wrapped to handle interruptions
func Interruptible(signals []os.Signal, worker goasync.Worker) *interruptible {
	return &interruptible{
		signals:  signals,
		callback: worker,
	}
}

func (i *interruptible) Initialize() error {
	return i.callback.Initialize()
}

func (i *interruptible) Cleanup() error {
	return i.callback.Cleanup()
}

func (i *interruptible) Work(ctx context.Context) error {
	for {
		workerCtx, _ := signal.NotifyContext(ctx, i.signals...)
		if err := i.callback.Work(workerCtx); err != nil {
			// if there is an error, always return it
			return err
		}

		select {
		case <-ctx.Done():
			// we are shutting down the application and don't have an error
			return nil
		default:
			select {
			case <-workerCtx.Done():
				// This is the case where the Interupt signal triggered the context to close and not the parent.
				// So we need to restart our worker
				if err := i.callback.Cleanup(); err != nil {
					return err
				}
				if err := i.callback.Initialize(); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Interruptible worker unexpectedly stopped")
			}
		}
	}
}
