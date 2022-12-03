package goasync

type Config struct {
	// When set to 'true' and Task.Execute() returning will cause the task mangager to shut down.
	// When set to 'false' a Task.Execute() returning will not cause the task manager to shut down. It
	// will only shutdown when the original context is canceled.
	AllowExecuteFailures bool

	// Paramter is only used when AllowExecuteFailures = true.
	//	If set to true - will hold on to and report errors when context is canceled
	//  If set to false - will not report any errors when context is canceled
	ReportErrors bool
}

// Default configuration that will cause an error if any tasks stop executing
func StrictConfig() Config {
	return Config{
		AllowExecuteFailures: false,
		ReportErrors:         true,
	}
}

// Default configuration that allows tasks stop executing without a failure
// and no errors will be reported
func RelaxedConfig() Config {
	return Config{
		AllowExecuteFailures: false,
		ReportErrors:         false,
	}
}
