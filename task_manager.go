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

	runningOnce *sync.Once
	running     chan struct{}

	tasksCounter    int
	stopTaskCounter int
	addTaskChan     chan struct{}
	removeTaskChan  chan struct{}

	// chan any tasks that are running can report an error on
	namedErrorChan chan *NamedError

	// context created after Run() is called. This is used to manage all threads
	ctx context.Context

	taskLock     *sync.Mutex
	namedTasks   []namedTask
	executeTasks []executeTask
}

//	RETURNS:
//	- *TaskManager - task manager with all private fields setup
//
// Create a new TaskManger to manage async tasks
func NewTaskManager() *TaskManager {
	return &TaskManager{
		doneOnce: new(sync.Once),
		done:     make(chan struct{}),

		runningOnce: new(sync.Once),
		running:     make(chan struct{}),

		tasksCounter:    0,
		stopTaskCounter: 0,
		addTaskChan:     make(chan struct{}),
		removeTaskChan:  make(chan struct{}),

		namedErrorChan: make(chan *NamedError),

		taskLock:     new(sync.Mutex),
		namedTasks:   make([]namedTask, 0),
		executeTasks: make([]executeTask, 0),
	}
}

//	PARAMETERS:
//	- name - name of the task. On any errors this will be reported to see which tasks failed
//	- task - task to be managed, cannot be nil
//	- taskType - type of the task to run and manage. This indicate the error handling and shutdown logic for the specific task
//
//	RETURNS
//	- error - any errors when adding the task to be managed
//
// Add a task to the TaskManager. All tasks added this way must be called
// before the Run(...) function so their Inintialize() function can be called in the proper order
func (t *TaskManager) AddTask(name string, task Task, taskType TASK_TYPE) error {
	t.taskLock.Lock()
	defer t.taskLock.Unlock()

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	select {
	case <-t.done:
		// check done before running
		return fmt.Errorf("task Manager has already shutdown. Will not manage task")
	default:
		select {
		case <-t.running:
			return fmt.Errorf("task Manager is already running. Will not manage task")
		default:
			t.namedTasks = append(t.namedTasks, namedTask{name: name, task: task, taskType: taskType})
			return nil
		}
	}
}

//	PARAMETERS:
//	- name - name of the task. On any errors this will be reported to see which tasks failed
//	- task - task to be managed, cannot be nil
//	- taskType - type of the task to run and manage. This indicate the error handling and shutdown logic for the specific task
//
//	RETURNS
//	- error - any errors when adding the task to be managed
//
// AddExecuteTask can be used to add a new task to the Taskmanger before or after it is already running
func (t *TaskManager) AddExecuteTask(name string, task ExecuteTask, taskType EXECUTE_TASK_TYPE) error {
	t.taskLock.Lock()
	defer t.taskLock.Unlock()

	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	select {
	case <-t.done:
		return fmt.Errorf("task Manager has already shutdown. Will not manage task")
	default:
		select {
		case <-t.running:
			// Add a new execute to the already running task manager
			select {
			case t.addTaskChan <- struct{}{}:
				go func() {
					t.namedErrorChan <- &NamedError{TaskName: name, Stage: Execute, Err: task.Execute(t.ctx), TaskType: taskType}
				}()
			case <-t.done:
				// must be shutting down There is a slim race when trying to add tasks async and a cancel happen at the same time
				return fmt.Errorf("task Manager has already shutdown. Will not manage task")
			}
		default:
			t.executeTasks = append(t.executeTasks, executeTask{name: name, task: task, taskType: taskType})
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
//	 2. In Parallel, run all Execute(...) functions for any tasks
//	    a. All tasks are expected to run and only allow nil returns if the configration allows for it.
//	    b. If any tasks return an error, the TaskManager can cancel all running tasks and then run the Cleanup for each task
//	 3. Once the context ic canceled, each task process will have their context canceled
//	 4. Each Task's Cleanup function is called in reverse order they were added to the TaskManager
//
//	Rules for adding tasks (added after Run):
//	 1. These tasks will also cause the Task Manager to shutdown if an error is encountered depending on configuration
func (t *TaskManager) Run(ctx context.Context) []NamedError {
	var taskErrors []NamedError
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

			// initialize all named tasks
			for index, namedTask := range t.namedTasks {
				if err := namedTask.task.Initialize(taskCtx); err != nil {
					taskErrors = append(taskErrors, NamedError{TaskName: namedTask.name, Stage: Initialize, Err: err})

					// we hit an error. Run Cleanup in reverse order
					for i := index; i >= 0; i-- {
						if err = t.namedTasks[i].task.Cleanup(taskCtx); err != nil {
							taskErrors = append(taskErrors, NamedError{TaskName: t.namedTasks[i].name, Stage: Cleanup, Err: err})
						}
					}

					return taskErrors
				}
			}

			// start all tasks
			for _, namedWork := range t.namedTasks {
				t.tasksCounter += 1

				go func(namedTask namedTask) {
					t.namedErrorChan <- &NamedError{TaskName: namedTask.name, Stage: Execute, Err: namedTask.task.Execute(taskCtx), TaskType: namedTask.taskType}
				}(namedWork)
			}

			// start all execute tasks
			for _, executeWork := range t.executeTasks {
				t.tasksCounter += 1

				go func(executeTask executeTask) {
					t.namedErrorChan <- &NamedError{TaskName: executeTask.name, Stage: Execute, Err: executeTask.task.Execute(taskCtx), TaskType: executeTask.taskType}
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
						}
					case <-t.addTaskChan: // signal a process is adding a task
						t.tasksCounter += 1
					case <-t.removeTaskChan: // signal a process is removing a task
						t.tasksCounter -= 1
					}
				}
			}()

		RUNLOOP:
			for {
				select {
				case namedError := <-t.namedErrorChan:
					t.removeTaskChan <- struct{}{} // signal that we are removing a task

					// record any possible errors
					if namedError.Err != nil {
						taskErrors = append(taskErrors, *namedError)
					}

					switch namedError.TaskType {
					case TASK_TYPE_STRICT:
						select {
						case <-taskCtx.Done():
							// nothing to do here
						default:
							cancel()

							// record unexpected return from an async func
							if namedError.Err == nil {
								namedError.Err = unexpectedStop
								taskErrors = append(taskErrors, *namedError)
							}
						}
					case TASK_TYPE_ERROR:
						// if the error is anything other than nil, we need to cancel all other running tasks
						if namedError.Err != nil {
							cancel()
						}
					case TASK_TYPE_STOP_GROUP:
						// if the error is anything other than nil, we need to cancel all other running tasks
						if namedError.Err != nil {
							cancel()
						}

						// all stop group tests have finished
						t.stopTaskCounter--
						if t.stopTaskCounter <= 0 {
							cancel()
						}
					default:
						panic("unknown task type to process error")
					}
				case <-t.done:
					// in this case, we have stopped processing all our background processes so we can exit
					break RUNLOOP
				}
			}

			// cleanup
			for i := len(t.namedTasks) - 1; i >= 0; i-- {
				if err := t.namedTasks[i].task.Cleanup(taskCtx); err != nil {
					taskErrors = append(taskErrors, NamedError{TaskName: t.namedTasks[i].name, Stage: Cleanup, Err: err})
				}
			}

			return taskErrors
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
