package goasync

import (
	"context"
	"fmt"
	"sync"
)

type taskManager struct {
	doneOnce *sync.Once
	done     chan struct{}

	cfg Config

	runningOnce *sync.Once
	running     chan struct{}

	// chan any tasks that are running can report an error on
	namedErrorChan chan NamedError

	// context created after Run() is called. This is used to manage all threads
	ctx context.Context

	tasks        *sync.WaitGroup
	taskLock     *sync.Mutex
	namedTasks   []namedTask
	executeTasks []executeTask
}

// Create a new task manager for async tasks
//
// PARAMS:
// * config - configuration to use. Handles error reporting and use cases for terminating the managed tasks
//
// RETURNS:
// * *taskManager - configured task manager
func NewTaskManager(config Config) *taskManager {
	return &taskManager{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		cfg: config,

		runningOnce: new(sync.Once),
		running:     make(chan struct{}),

		namedErrorChan: make(chan NamedError),

		tasks:    new(sync.WaitGroup),
		taskLock: new(sync.Mutex),
	}
}

// Add a task to the TaskManager. All tasks added this way must be called
// before the Run(...) function so their Inintialize() function can be called properly
//
// PARAMS:
// * name - name of the task. On any errors this will be reported with the name
// * task - task to be managed
//
// REETURNS
// * error - any errors when adding the task to be managed
func (t *taskManager) AddTask(name string, task Task) error {
	t.taskLock.Lock()
	defer t.taskLock.Unlock()

	select {
	case <-t.done:
		// check done before running
		return fmt.Errorf("Task Manager has already shutdown. Will not manage task")
	default:
		select {
		case <-t.running:
			return fmt.Errorf("Task Manager is already running. Will not manage task")
		default:
			t.namedTasks = append(t.namedTasks, namedTask{name: name, task: task})
			return nil
		}
	}
}

// AddExecuteTask can be used to add a new task to the Taskmanger before or after it is already running
//
// PARAMS:
// * name - name of the task. On any errors this will be reported with the name
// * task - task to be managed
//
// REETURNS
// * error - any errors when adding the task to be managed
func (t *taskManager) AddExecuteTask(name string, task ExecuteTask) error {
	t.taskLock.Lock()
	defer t.taskLock.Unlock()

	select {
	case <-t.done:
		return fmt.Errorf("Task Manager has already shutdown. Will not manage task")
	default:
		select {
		case <-t.running:
			// Add a new threaded task
			t.tasks.Add(1)
			go func() {
				defer t.tasks.Done()
				err := task.Execute(t.ctx)

				if err != nil {
					select {
					case t.namedErrorChan <- NamedError{TaskName: name, Stage: Execute, Err: err}:
					case <-t.done:
						// There is a very small race condition here where this might not be added. It can happen
						// if a task fails at the same time this is added to be managed. So just be safe and cleanup
						// this possible leaked thread and the error message will be lost. But I think thats fine since
						// it wasn't the thread that caused the initial failure
					}
				} else {
					select {
					case <-t.done:
						// nothing to do here since we are properly shutting down
					case t.namedErrorChan <- NamedError{TaskName: name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}:
						// unexpected shutdown for a task process. Initiate abort of all task processes
					}
				}
			}()
		default:
			t.executeTasks = append(t.executeTasks, executeTask{name: name, task: task})
		}
	}

	return nil
}

// Run any tasks added to the TaskManager.
//
// Rules for Tasks (added before Run):
//  1. Run each Initialize process serially in the order they were added to the TaskManager
//     a. If an error occurs, stop Initializng any remaning tasks. Also Run Cleanup for
//     any tasks that have been Initialized
//  2. In Parallel Run all Execute(...) functions for any tasks
//     a. All tasks are expected to run and not error.
//     b. If any tasks return an error, the TaskManager will cancel all running tasks and then
//     run the Cleanup for each task if configured to do so.
//  3. Once the context ic canceled, each task process will have their context canceled
//  4. Each Task's Cleanup function is called in reverse order they were added to the TaskManager
//
// Rules for adding tasks (added after Run):
//  1. These tasks will also cause the Task Manager to shutdown if an error is encountered if configured to do so
//
// PARAMS:
// * ctx - context to stop the TaskManager
//
// RETURNS:
// * []NamedError - slice of errors with the name of the failed task
func (t *taskManager) Run(ctx context.Context) []NamedError {
	var errors []NamedError

	t.taskLock.Lock()
	select {
	case <-t.done:
		// don't re-run if already stopped. closed channels will cause a problem
		t.taskLock.Unlock()
		return []NamedError{
			{Err: fmt.Errorf("task manager has already shutdown. Will not run again")},
		}
	default:
		select {
		case <-t.running:
			// don't re-run if already stopped. closed channels will cause a problem
			t.taskLock.Unlock()
			return []NamedError{
				{Err: fmt.Errorf("task manager is already running. Will not run again")},
			}
		default:
			// setup for AddExecuteTask to ensure that can run properly
			taskCtx, cancel := context.WithCancel(ctx)
			t.ctx = taskCtx
			t.closeRunning()
			t.taskLock.Unlock()

			// initialize all named tasks
			for index, namedTask := range t.namedTasks {
				if err := namedTask.task.Initialize(); err != nil {
					errors = append(errors, NamedError{TaskName: namedTask.name, Stage: Initialize, Err: err})

					// we hit an error. Run Cleanup in reverse order
					for i := index; i >= 0; i-- {
						if err = t.namedTasks[i].task.Cleanup(); err != nil {
							errors = append(errors, NamedError{TaskName: t.namedTasks[i].name, Stage: Cleanup, Err: err})
						}
					}

					cancel()
					return errors
				}
			}

			// start all tasks
			for _, namedWork := range t.namedTasks {
				t.tasks.Add(1)
				go func(namedTask namedTask) {
					defer t.tasks.Done()

					err := namedTask.task.Execute(taskCtx)
					if err != nil {
						t.namedErrorChan <- NamedError{TaskName: namedTask.name, Stage: Execute, Err: err}
					} else {
						select {
						case <-taskCtx.Done():
							// nothing to do here since we are properly shutting down
						default:
							// unexpected shutdown for a task process. Initiate abort of all task processes
							t.namedErrorChan <- NamedError{TaskName: namedTask.name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}
						}
					}
				}(namedWork)
			}

			// start all execute tasks
			for _, executeWork := range t.executeTasks {
				t.tasks.Add(1)
				go func(executeTask executeTask) {
					defer t.tasks.Done()

					err := executeTask.task.Execute(taskCtx)
					if err != nil {
						t.namedErrorChan <- NamedError{TaskName: executeTask.name, Stage: Execute, Err: err}
					} else {
						select {
						case <-taskCtx.Done():
							// nothing to do here since we are properly shutting down
						default:
							// unexpected shutdown for a task process. Initiate abort of all task processes
							t.namedErrorChan <- NamedError{TaskName: executeTask.name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}
						}
					}
				}(executeWork)
			}

			go func() {
				if t.cfg.AllowNoManagedProcesses {
					// need to wait for the original context to close. Or a failure is  shutting down
					// the server and we need to shutdown
					<-taskCtx.Done()
				}

				// in both cases, we still want for all tasks to finish draining
				t.tasks.Wait()
				t.closeDone()
			}()

		RUNLOOP:
			for {
				select {
				case namedError := <-t.namedErrorChan:
					if t.cfg.AllowExecuteFailures {
						// we are allowing execute failures, do don't cancel the task manager
						if t.cfg.ReportErrors {
							errors = append(errors, namedError)
						}
					} else {
						// any error should be recored and fail the task manger since we do not allow for any errors
						errors = append(errors, namedError)
						cancel()
					}
				case <-t.done:
					// in this case, we have stopped processing all our background processes so we can exit
					break RUNLOOP
				}
			}

			// cleanup
			for i := len(t.namedTasks) - 1; i >= 0; i-- {
				if err := t.namedTasks[i].task.Cleanup(); err != nil {
					errors = append(errors, NamedError{TaskName: t.namedTasks[i].name, Stage: Cleanup, Err: err})
				}
			}

			return errors
		}
	}
}

func (t *taskManager) closeDone() {
	t.doneOnce.Do(func() {
		close(t.done)
	})
}

func (t *taskManager) closeRunning() {
	t.runningOnce.Do(func() {
		close(t.running)
	})
}
