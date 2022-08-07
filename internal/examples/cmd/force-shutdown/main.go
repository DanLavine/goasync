package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/internal/examples/pkg/forceshutdown"
	"github.com/DanLavine/goasync/tasks"
)

// This is an example of a bad acting sub process that won't shut down by default.
// This setup ensures our program always exits

func main() {
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)

	foreverTask := forceshutdown.ForeverTask()

	taskManger := goasync.NewTaskManager()
	taskManger.AddTask("foreverTask", tasks.ForceStop(5*time.Second, foreverTask))

	if errs := taskManger.Run(shutdown); errs != nil {
		log.Fatal(errs)
	}
}
