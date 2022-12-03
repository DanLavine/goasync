# GoAsync - Flexible processes management library

Go Async is designed to simplify managing multi-process applications into easily understandable
single purpose units of work.

There `Task Manager` is responsible for running and managing any number of assigned `tasks`.
For the Task Manager to operate, each task must follow a specific number of rules:

1. Each task added to the Task Manager will initialize serially in the order they were added to the Task Manager
  a. If an error occurs, stop Initializng any remaning tasks. Also Run Cleanup for any tasks that have already been Initialized
2. In Parallel Run all Execute(...) functions for any tasks
  a. All tasks are expected to run and not error.
  b. If any tasks return an error, the Task Manager will cancel all running tasks and then run the Cleanup for each task.
3. When the Task Manager's context is canceled, each task process will have their context canceled
4. Each Task's Cleanup function is called in reverse order they were added to the Task Manager

Lastly any tasks can be added to the Task Manager as long as they follow the Task interface:
```
type Task interface {
	// Initializate functions are ran serially in the order they were added to the TaskManager.
	// These are useful when one Goroutine dependency requires a previous Worker to setup some common
	// dependency like a DB connection.
	Initialize() error

	// Execute is the main Async function to house all the multi-threaded logic handled by GoAsync.
	Execute(ctx context.Context) error

	// Clenup functions are ran serially in reverse order they were added to the Task Manager.
	// This way the 1st Initialze dependency is stopped last
	Cleanup() error
}
```

For a number of common basic tasks provided, see [tasks](./tasks)


# Installation

```
go get -u github.com/DanLavine/goasync
```

# Examples

For a few simple and basic real world example check out the `internal/examples` dir

# Running tests

To run all tests:
```
go test --race ./...
```

If we want to run one specific test we can use:
```
go test --race -run [NAME OF TEST] ./...
```

To ignore the test cache:
```
go test --race -count=1 ./...
```
