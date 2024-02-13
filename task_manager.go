package goasync

import (
	"context"
	"fmt"
	"sync"
)

// TaskManager implements the AsyncTaskManager interface and contains the logic for managing any Tasks
// proovided to this struct.
type TaskManager struct {
	doneOnce *sync.Once
	done     chan struct{}

	cfg Config

	runningOnce *sync.Once
	running     chan struct{}

	tasksCounter   int
	addTaskChan    chan struct{}
	removeTaskChan chan struct{}

	// chan any tasks that are running can report an error on
	namedErrorChan chan *NamedError

	// context created after Run() is called. This is used to manage all threads
	ctx context.Context

	taskLock     *sync.Mutex
	namedTasks   []namedTask
	executeTasks []executeTask
}

//	PARAMETERS:
//	- config - configuration to use. Handles error reporting and use cases for terminating the managed tasks
//
//	RETURNS:
//	- *TaskManager - configured task manager
//
// Create a new TaskManger to manage async tasks
func NewTaskManager(config Config) *TaskManager {
	return &TaskManager{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		cfg: config,

		runningOnce: new(sync.Once),
		running:     make(chan struct{}),

		tasksCounter:   0,
		addTaskChan:    make(chan struct{}),
		removeTaskChan: make(chan struct{}),

		namedErrorChan: make(chan *NamedError),

		taskLock: new(sync.Mutex),
	}
}

//	PARAMETERS:
//	- name - name of the task. On any errors this will be reported to see which tasks failed
//	- task - task to be managed, cannot be nil
//
//	RETURNS
//	- error - any errors when adding the task to be managed
//
// Add a task to the TaskManager. All tasks added this way must be called
// before the Run(...) function so their Inintialize() function can be called in the proper order
func (t *TaskManager) AddTask(name string, task Task) error {
	t.taskLock.Lock()
	defer t.taskLock.Unlock()

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

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

//	PARAMETERS:
//	- name - name of the task. On any errors this will be reported to see which tasks failed
//	- task - task to be managed, cannot be nil
//
//	RETURNS
//	- error - any errors when adding the task to be managed
//
// AddExecuteTask can be used to add a new task to the Taskmanger before or after it is already running
func (t *TaskManager) AddExecuteTask(name string, task ExecuteTask) error {
	t.taskLock.Lock()
	defer t.taskLock.Unlock()

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	select {
	case <-t.done:
		return fmt.Errorf("Task Manager has already shutdown. Will not manage task")
	default:
		select {
		case <-t.running:
			// Add a new threaded task
			select {
			case t.addTaskChan <- struct{}{}:
				// // can still add a task
				// t.addTaskChan <- struct{}{} // singnal that we are done adding the task

				go func() {
					err := task.Execute(t.ctx)

					if err != nil {
						t.namedErrorChan <- &NamedError{TaskName: name, Stage: Execute, Err: err}
					} else {
						select {
						case <-t.done:
							// nothing to do here since we are properly shutting down and want to capture the race
							t.namedErrorChan <- nil
						case t.namedErrorChan <- &NamedError{TaskName: name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}:
							// unexpected shutdown for a task process. Initiate abort of all task processes
						}
					}
				}()
			case <-t.done:
				// must be shutting down There is a slim rce when trying to add tasks async and a cancel happen at the same time
				return fmt.Errorf("Task Manager has already shutdown. Will not manage task")
			}
		default:
			t.executeTasks = append(t.executeTasks, executeTask{name: name, task: task})
		}
	}

	return nil
}

//	PARAMETERS:
//	- ctx - context to stop the TaskManager
//
//	RETURNS:
//	- []NamedError - slice of errors with the name of the failed task
//
// Run any tasks added to the TaskManager.
//
//	Rules for Tasks (added before Run):
//	 1. Run each Initialize process serially in the order they were added to the TaskManager
//	    a. If an error occurs, stop Initializng any remaning tasks. Also Run Cleanup for any tasks that have been Initialized.
//	 2. In Parallel Run all Execute(...) functions for any tasks
//	    a. All tasks are expected to run and only error if the configration allows for it.
//	    b. If any tasks return an error, the TaskManager can cancel all running tasks and then run the Cleanup for each task if configured to do so.
//	 3. Once the context ic canceled, each task process will have their context canceled
//	 4. Each Task's Cleanup function is called in reverse order they were added to the TaskManager
//
//	Rules for adding tasks (added after Run):
//	 1. These tasks will also cause the Task Manager to shutdown if an error is encountered depending on configuration
func (t *TaskManager) Run(ctx context.Context) []NamedError {
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
			// don't re-run if already running
			t.taskLock.Unlock()
			return []NamedError{
				{Err: fmt.Errorf("task manager is already running. Will not run again")},
			}
		default:
			// setup for AddExecuteTask to ensure that can run properly
			taskCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			t.ctx = taskCtx
			t.closeRunning()
			t.taskLock.Unlock()

			if len(t.namedTasks) == 0 && len(t.executeTasks) == 0 && !t.cfg.AllowNoManagedProcesses {
				t.closeDone()
				return []NamedError{
					{Err: fmt.Errorf("no tasks were added to the task manager and the configuration has AllowNoManagedProcesses = false")},
				}
			}

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

					return errors
				}
			}

			// start all tasks
			for _, namedWork := range t.namedTasks {
				t.tasksCounter += 1

				go func(namedTask namedTask) {
					err := namedTask.task.Execute(taskCtx)
					if err != nil {
						t.namedErrorChan <- &NamedError{TaskName: namedTask.name, Stage: Execute, Err: err}
					} else {
						select {
						case <-taskCtx.Done():
							// shutting down
							t.namedErrorChan <- nil
						default:
							// unexpected shutdown for a task process. Initiate abort of all task processes
							t.namedErrorChan <- &NamedError{TaskName: namedTask.name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}
						}
					}
				}(namedWork)
			}

			// start all execute tasks
			for _, executeWork := range t.executeTasks {
				t.tasksCounter += 1

				go func(executeTask executeTask) {
					err := executeTask.task.Execute(taskCtx)
					if err != nil {
						t.namedErrorChan <- &NamedError{TaskName: executeTask.name, Stage: Execute, Err: err}
					} else {
						select {
						case <-taskCtx.Done():
							// shutting down
							t.namedErrorChan <- nil
						default:
							// unexpected shutdown for a task process. Initiate abort of all task processes
							t.namedErrorChan <- &NamedError{TaskName: executeTask.name, Stage: Execute, Err: fmt.Errorf("Unexpected shutdown for task process")}
						}
					}
				}(executeWork)
			}

			// manage threading for tring to add a task and shutdown race conditions
			go func() {
				for {
					select {
					case <-taskCtx.Done():
						// the main context was canceld by either a shutdown from the caller or an error
						for {
							if t.tasksCounter == 0 {
								t.closeDone()
								return
							}

							<-t.removeTaskChan
							t.tasksCounter -= 1

							if !t.cfg.AllowNoManagedProcesses {
								// if there are no managed proccesses we are done
								if t.tasksCounter == 0 {
									t.closeDone()
								}

								// must still be processing shutdowns so nothing to do
							} else {
								select {
								case <-taskCtx.Done():
									// the task context is done, so we are canceling or there was an error
									if t.tasksCounter == 0 {
										t.closeDone()
									}
								default:
									// must still be processing shutdowns so nothing to do
								}
							}
						}
					case <-t.addTaskChan: // signal a process is adding a task
						t.tasksCounter += 1
					case <-t.removeTaskChan:
						t.tasksCounter -= 1
					}
				}
			}()

		RUNLOOP:
			for {
				select {
				case namedError := <-t.namedErrorChan:
					t.removeTaskChan <- struct{}{} // signal that we are removing a task

					if t.cfg.AllowExecuteFailures {
						// we are allowing execute failures, do don't cancel the task manager
						if t.cfg.ReportErrors {
							if namedError != nil {
								errors = append(errors, *namedError)
							}
						}
					} else {
						// any error should be recored and fail the task manger since we do not allow for any errors
						if t.cfg.ReportErrors {
							if namedError != nil {
								errors = append(errors, *namedError)
							}
						}

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

func (t *TaskManager) closeDone() {
	t.doneOnce.Do(func() {
		close(t.done)
	})
}

func (t *TaskManager) closeRunning() {
	t.runningOnce.Do(func() {
		close(t.running)
	})
}
