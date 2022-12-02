package goasync

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// A Task is anything that can be managed by the TaskManager and added before the taskmanager
// start running. Any errors will cause the process to exit as all tasks are expected to run without erros
//
//counterfeiter:generate . Task
type Task interface {
	// Initializate functions are ran serially in the order they were added to the TaskManager.
	// These are useful when one Goroutine dependency requires a previous Worker to setup some common
	// dendency like a DB connection.
	Initialize() error

	// Execute is the main Async function to house all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error

	// Clenup functions are ran serially in reverse order they were added to the TaskManager.
	// This way the 1st Initialze dependency is stopped last
	Cleanup() error
}

// A RunningTask can be added to a Task Manager after it has already started managin the tasks.
// These tasks are expected to already be properly Initialized and don't require any Cleanup Code.
type RunningTask interface {
	// Execute is the main Async function to house all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error
}

type namedTask struct {
	name string
	task Task
}
