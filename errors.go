package goasync

import (
	"fmt"
)

type Stage string

var (
	unexpectedStop = fmt.Errorf("unexpected shutdown for task process")
)

const (
	Initialize Stage = "initialize"
	Execute    Stage = "execute"
	Cleanup    Stage = "cleanup"
)

type NamedError struct {
	TaskName string
	Stage    Stage
	Err      error
	TaskType TASK_TYPE
}
