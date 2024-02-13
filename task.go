package goasync

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// AsyncTaskManager manages any number of tasks
//
//counterfeiter:generate . AsyncTaskManager
type AsyncTaskManager interface {
	// Add a task before running the task manager
	AddTask(name string, task Task) error

	// add a task to an already running task manager
	AddExecuteTask(name string, task ExecuteTask) error

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
	Initialize() error

	// Execute is the main Async function to contain all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error

	// Cleanup functions are ran serially in reverse order they were added to the TaskManager.
	// This way the 1st Initalized dependency is stopped last
	Cleanup() error
}

// ExecuteTask can be added to the TaskManager before or after it has already started Executing the tasks.
// These tasks are expected to already be properly Initialized and don't require any Cleanup Code.
//
//counterfeiter:generate . ExecuteTask
type ExecuteTask interface {
	// Execute is the main Async function to contain all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error
}

type namedTask struct {
	name string
	task Task
}

type executeTask struct {
	name string
	task ExecuteTask
}
