package workers

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/DanLavine/goasync"
)

type interruptible struct {
	interrupt chan os.Signal
	callback  goasync.Worker
}

// Interruptible Workers are used to run any tasks that can be reloaded at runtime.
// When this worker recieves an interrupt signal it will cancel the child Worker process
// and call Cleanup. Then this will call Initialize and Work again for the child Worker.
//
// Args:
//  - signals: OS signals to capture and initiate the reload operation
//  - worker: worker process that is wrapped to handle interruptions
func Interruptible(signals []os.Signal, worker goasync.Worker) *interruptible {
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, signals...)

	return &interruptible{
		interrupt: interrupt,
		callback:  worker,
	}
}

func (i *interruptible) Initialize() error {
	return i.callback.Initialize()
}

func (i *interruptible) Cleanup() error {
	return i.callback.Cleanup()
}

func (i *interruptible) Work(ctx context.Context) error {
	restarting := false
	errChan := make(chan error)
	defer close(errChan)

	// start running the inital process
	workerCtx, cancel := context.WithCancel(ctx)
	go func(workerCtx context.Context) {
		errChan <- i.callback.Work(workerCtx)
	}(workerCtx)

	for {
		select {
		case <-i.interrupt:
			// we have recieved an interupt.
			cancel()
			restarting = true
		case err := <-errChan:
			// if there is an error, always return it
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				// we are shutting down the application and don't have an error
				return nil
			default:
				// if we are not restarting, then the child woker unxepectidly returned
				fmt.Println("restarting?", restarting)
				if !restarting {
					return fmt.Errorf("Interruptible worker unexpectedly stopped")
				}

				// Cleanup and re-Initialize
				if err := i.callback.Cleanup(); err != nil {
					return err
				}
				if err := i.callback.Initialize(); err != nil {
					return err
				}

				// renew our context and start running again
				workerCtx, cancel = context.WithCancel(ctx)

				// run worker in background
				go func(workerCtx context.Context) {
					errChan <- i.callback.Work(workerCtx)
				}(workerCtx)

				restarting = false
			}
		}
	}
}
