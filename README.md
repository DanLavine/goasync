# GoAsync - Flexible processes management library
[godoc](https://pkg.go.dev/github.com/DanLavine/goasync)

Go Async is designed to simplify managing multi-process applications into easily understandable
single purpose units of work.

The `Task Manager` is responsible for running and managing any number of assigned `tasks`.
For the Task Manager to operate, each task must follow 1 of two patterns.

### Long Running Task
These tasks are expected to run for the entire task manager lifecyle as they have special initialization
and shutdown logic. 

1. Each task added to the Task Manager will initialize serially in the order they were added to the Task Manager
    1. If an error occurs during Initialization, any remaining tasks are skipped and Cleanup is called for any tasks that have already been Initialized
2. In parallel runs all Execute(...) functions for any tasks
    1. If the task manager is configured to abort execution on any stoped task, then the Task Manager will
       cancel all running tasks and then run the Cleanup for each task.
3. When the Task Manager's context is canceled, each task process will have their context canceled. The task manager will then wait until
   all running tasks have been stopped.
4. Each Task's Cleanup function is called in reverse order they were added to the Task Manager

Task interface:
```
// Task is anything that can be added to the TaskManager before it stats Executing.
type Task interface {
// Initialize functions are ran serially in the order they were added to the TaskManager.
// These are useful when one Goroutine dependency requires a previous Worker to setup some common
// dependency like a DB connection.
Initialize() error

// Execute is the main Async function to house all the multi-threaded logic handled by GoAsync.
Execute(ctx context.Context) error

// Cleanup functions are ran serially in reverse order they were added to the TaskManager.
// This way the 1st Initalized dependency is stopped last
Cleanup() error
}
```

### Execute Task
Can be added to a process that is already running

ExecuteTask interface:
```
// ExecuteTask can be added to the TaskManager before or after it has already started Executing the tasks.
// These tasks are expected to already be properly Initialized and don't require any Cleanup Code.
type ExecuteTask interface {
  // Execute is the main Async function to house all the multi-threaded logic handled by GoAsync.
  Execute(ctx context.Context) error
}
```

# Provided tasks
For a number of common basic tasks provided, see [tasks](./tasks) directory

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
