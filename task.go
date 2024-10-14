package goasync

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// AsyncTaskManager manages any number of tasks
//
//counterfeiter:generate . AsyncTaskManager
type AsyncTaskManager interface {
	// Add a task before running the task manager
	AddTask(name string, task Task, taskType TASK_TYPE) error

	// add a task to an already running task manager
	AddExecuteTask(name string, task ExecuteTask, taskType EXECUTE_TASK_TYPE) error

	// run the task manager
	Run(context context.Context) []NamedError
}

// Task is anything that can be added to the TaskManager before it stats Executing.
//
//counterfeiter:generate . Task
type Task interface {
	// Initialize functions are ran serially in the order they were added to the AsyncTaskManager.
	// These are useful when one go routine dependency requires a previous Worker to setup some common
	// dendency like a DB connection.
	Initialize(ctx context.Context) error

	// Execute is the main Async function to contain all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error

	// Cleanup functions are ran serially in reverse order they were added to the TaskManager.
	// This way the 1st Initalized dependency is stopped last
	Cleanup(ctx context.Context) error
}

// ExecuteTask can be added to the TaskManager before or after it has already started Executing the tasks.
// These tasks are expected to already be properly Initialized and don't require any Cleanup Code.
//
//counterfeiter:generate . ExecuteTask
type ExecuteTask interface {
	// Execute is the main Async function to contain all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error
}

// TASK_TYPE defines how tasks should be handeled when they stop processing
type TASK_TYPE int

const (
	// STRICT will cause a shutdown to all other tasks if they return before the Task Manager's Context has been canceled
	TASK_TYPE_STRICT TASK_TYPE = iota
	// ERROR will cause a shutdown to all other tasks if any non nil error is returned before the Task Manager's Context has been canceled
	TASK_TYPE_ERROR
	// STOP_GROUP will cause a shutdown on the Task Manager once all STOP_GROUP process have finished iff they only return nil error(s).
	// If any Task(s) returns an error that is not nil, then the Task Manager will be canceled and all errors will be reported
	TASK_TYPE_STOP_GROUP
)

// EXECUTE_TASK_TYPE defines the possible task types that can applied to an already running Task Manager
type EXECUTE_TASK_TYPE = TASK_TYPE

const (
	// STRICT will cause a shutdown to all other tasks if they return before the Task Manager's Context has been canceled
	EXECUTE_TASK_TYPE_STRICT EXECUTE_TASK_TYPE = TASK_TYPE_STRICT

	// ERROR will cause a shutdown to all other tasks if any non nil error is returned before the Task Manager's Context has been canceled
	EXECUTE_TASK_TYPE_ERROR EXECUTE_TASK_TYPE = TASK_TYPE_ERROR
)

type namedTask struct {
	name string
	task Task

	taskType TASK_TYPE
}

type executeTask struct {
	name string
	task ExecuteTask

	taskType EXECUTE_TASK_TYPE
}
