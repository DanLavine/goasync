package forceshutdown

import (
	"context"
	"fmt"
	"time"
)

type foreverTask struct{}

func ForeverTask() foreverTask {
	return foreverTask{}
}

func (f foreverTask) Initialize() error { return nil }
func (f foreverTask) Cleanup() error    { return nil }

func (f foreverTask) Execute(ctx context.Context) error {
	for {
		fmt.Println("Never going to stop printing!")
		time.Sleep(time.Second)
	}

	return nil
}
