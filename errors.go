package goasync

type Stage string

const (
	Initialize Stage = "initialize"
	Work       Stage = "work"
	Cleanup    Stage = "cleanup"
)

type NamedError struct {
	WorkerName string
	Stage      Stage
	Err        error
}
