package goasync

type Stage string

const (
	Initialize Stage = "initialize"
	Execute    Stage = "execute"
	Cleanup    Stage = "cleanup"
)

type NamedError struct {
	TaskName string
	Stage    Stage
	Err      error
}
