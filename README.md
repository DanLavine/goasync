# GoAsync - Flexible processes management library
[godoc](https://pkg.go.dev/github.com/DanLavine/goasync/v2)

Go Async is designed to simplify managing multi-process applications into easily understandable
single purpose units of work.

The `Task Manager` is responsible for running and managing any number of assigned `tasks`.
For the Task Manager to operate, each task must follow 1 of two patterns.

### Long running tasks
These tasks are intended to run for the majority of the Task Manager lifecycle and have dedicated
initialization and shutdown logic. 

1. Each task added to the Task Manager will initialize serially in the order they were added to the Task Manager
    1. If an error occurs during Initialization, any remaining tasks are skipped and Cleanup is called for any tasks that have already been Initialized
2. In parallel, all the Task's Execute(...) functions are ran.
    1. All tasks are then monitored for unexpected failures that will cause general shutdown of other tasks
    1. If special registered "shutdown tasks" finish processing, then all other tasks will be canceled gracefully 
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
  Initialize(ctx context.Context) error

  // Execute is the main Async function to house all the multi-threaded logic handled by GoAsync.
  Execute(ctx context.Context) error

  // Cleanup functions are ran serially in reverse order they were added to the TaskManager.
  // This way the 1st Initalized dependency is stopped last
  Cleanup(ctx context.Context) error
}
```

Adding Tasks:
```
// PARAMETERS:
// - name - name of the task. On any errors this will be reported to see which tasks failed
// - task - task to be managed, cannot be nil
// - taskType - type of the task to run and manage. This indicate the error handling and shutdown logic for the specific task
//
// RETURNS
// - error - any errors when adding the task to be managed
//
// Add a task to the TaskManager. All tasks added this way must be called
// before the Run(...) function so their Inintialize() function can be called in the proper order
AddTask(name string, task Task, taskType TASK_TYPE) error
```

When adding tasks, you can provide a number of `TASK_TYPE` values which dicate the logic for async management:
```
* TASK_TYPE_STRICT     - These tasks should only be stoped when the `Execute(ctx context.Context)` has a canceled ctx. If these error or return nil
                         early then all other tasks are stopped and the error is propigated through the Task Manager.
* TASK_TYPE_ERROR      - Will cause a shutdown to all other tasks if any non nil error is returned as part of the `Executes(ctx context.Context)` call.
* TASK_TYPE_STOP_GROUP - Stop Group tasks all need to be added before the Task Manager start to `Run(context.Context)` tasks and have special logic for
                         stopping all other tasks. These tasks are allowed to return nil errors and once all `TASK_TYPE_STOP_GROUP` have returned nil,
                         any other managed tasks are automatically canceled. If however, any of these async tasks return an actual error, all other task
                         will be stopped immediately and the error will be propigated.
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

Adding Tasks:
```
// PARAMETERS:
// - name - name of the task. On any errors this will be reported to see which tasks failed
// - task - task to be managed, cannot be nil
// - taskType - type of the task to run and manage. This indicate the error handling and shutdown logic for the specific task
//
// RETURNS
// - error - any errors when adding the task to be managed
//
// AddExecuteTask can be used to add a new task to the Taskmanger before or after it is already running
AddExecuteTask(name string, task ExecuteTask, taskType EXECUTE_TASK_TYPE) error
```

When adding tasks, you can provide a number of `TASK_TYPE` values which dicate the logic for async management:
```
* EXECUTE_TASK_TYPE_STRICT - These tasks should only be stoped when the `Execute(ctx context.Context)` has a canceled ctx. If these error or return nil
                             early then all other tasks are stopped and the error is propigated through the Task Manager.
* EXECUTE_TASK_TYPE_ERROR  - Will cause a shutdown to all other tasks if any non nil error is returned as part of the `Executes(ctx context.Context)` call.
```

# Provided tasks
For a number of common basic tasks provided, see [tasks](./tasks) directory

# Installation
```
go get -u github.com/DanLavine/goasync/v2
```

# Examples
For a few simple and basic real world example check out the `internal/examples` dir

# Running tests
To run all tests:
```
make test
```
