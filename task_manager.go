package goasync

import (
	"context"
	"fmt"
	"sync"
)

type taskManager struct {
	namedTasks []namedTask
}

func NewTaskManager() *taskManager {
	return &taskManager{}
}

// Add a task to the TaskManager
func (d *taskManager) AddTask(name string, task Task) {
	d.namedTasks = append(d.namedTasks, namedTask{name: name, task: task})
}

// Run any tasks added to the TaskManager.
//
// Rules for Tasks:
// 1. Run each Initialize process serially in the order they were added to the TaskManager
//   a. If an error occurs, stop Initializng any remaning tasks. Also Run Cleanup for
//      any already tasks that have been Initialized
// 2. In Parallel Run all Execute(...) functions for any tasks
//   a. All tasks are expected to run and not error.
//   b. If any tasks return an error, the TaskManager will cancel all running tasks and then
//      run the Cleanup for each task.
// 3. Once Stop is called for the TaskManager each task process will have their context canceled
// 4. Each Task's Cleanup function is called in reverse order they were added to the TaskManager
func (d *taskManager) Run(ctx context.Context) []NamedError {
	var errors []NamedError

	// initialize
	for index, namedTask := range d.namedTasks {
		if err := namedTask.task.Initialize(); err != nil {
			errors = append(errors, NamedError{TaskName: namedTask.name, Stage: Initialize, Err: err})

			// we hit an error. Run Cleanup in reverse order
			for i := index; i >= 0; i-- {
				if err = d.namedTasks[i].task.Cleanup(); err != nil {
					errors = append(errors, NamedError{TaskName: d.namedTasks[i].name, Stage: Cleanup, Err: err})
				}
			}

			return errors
		}
	}

	// tasks
	wg := &sync.WaitGroup{}
	finished := make(chan struct{})
	namedErrorChan := make(chan NamedError)
	taskCtx, cancel := context.WithCancel(ctx)

	// start all tasks
	for _, namedWork := range d.namedTasks {
		wg.Add(1)
		go func(namedTask namedTask) {
			defer wg.Done()

			err := namedTask.task.Execute(taskCtx)
			if err != nil {
				namedErrorChan <- NamedError{TaskName: namedTask.name, Stage: Execute, Err: err}
			} else {
				select {
				case <-taskCtx.Done():
					// nothing to do here since we are properly shutting down
				default:
					// unexpected shutdown for a task process. Initiate abort of all task processes
					namedErrorChan <- NamedError{TaskName: namedTask.name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}
				}
			}
		}(namedWork)
	}

	go func() {
		wg.Wait()
		close(finished)
	}()

RUNLOOP:
	for {
		select {
		case namedError := <-namedErrorChan:
			errors = append(errors, namedError)
			cancel()
		case <-finished:
			// in this case, we have stopped processing all our background processes so we can exit
			break RUNLOOP
		}
	}

	// cleanup
	for i := len(d.namedTasks) - 1; i >= 0; i-- {
		if err := d.namedTasks[i].task.Cleanup(); err != nil {
			errors = append(errors, NamedError{TaskName: d.namedTasks[i].name, Stage: Cleanup, Err: err})
		}
	}

	return errors
}
