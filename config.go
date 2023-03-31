package goasync

type Config struct {
	// When set to 'true' and TaskManager.Execute(...) returning will cause the task mangager to shut down.
	// When set to 'false' a TaskManager.Execute(...) returning will not cause the task manager to shut down. It
	// will only shutdown when the original context is canceled.
	AllowExecuteFailures bool

	// Paramter is only used when AllowExecuteFailures = true.
	//	If set to 'true', will hold on to and report errors when context is canceled
	//  If set to 'false', will not report any errors when context is canceled
	ReportErrors bool

	// When set tot 'true', this will allow the TaskManager to run even if no tasks are currently being managed. The use case
	// here is to allow for the the TaskManager.AddRunningTask(...) to periodically be called and
	// those process might just end. For example a TCP server that has clients connecting and disconnecting
	// can be used to add tasks here. But we would still want to cancel them when the original context
	// is told to close (aka shutdown the server)
	AllowNoManagedProcesses bool
}

// Default configuration that will cause an error if any tasks stop executing
func StrictConfig() Config {
	return Config{
		AllowExecuteFailures:    false,
		ReportErrors:            true,
		AllowNoManagedProcesses: false,
	}
}

// Default configuration that allows tasks stop executing without a failure
// and no errors will be reported
func RelaxedConfig() Config {
	return Config{
		AllowExecuteFailures:    true,
		ReportErrors:            false,
		AllowNoManagedProcesses: true,
	}
}
