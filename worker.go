package goasync

import "context"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Worker
type Worker interface {
	// Initializate functions are ran serially in the order they were added to the Director.
	// These are useful when one Goroutine dependency requires a previous Worker to setup some common
	// dendency like a DB connection.
	Initialize() error

	// Work is the main Async function to house all the multi-threaded logic handled by GoAsync.
	Work(ctx context.Context) error

	// Clenup functions are ran serially in reverse order they were added to the Director.
	// This way the 1st Initialze dependency is stopped last
	Cleanup() error
}

type namedWorker struct {
	name   string
	worker Worker
}
