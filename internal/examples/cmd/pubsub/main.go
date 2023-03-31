package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/DanLavine/goasync"
	"github.com/DanLavine/goasync/internal/examples/pkg/pubsub"
)

// This is an example of a Pub Sub system where on shutdown, all messages are
// finished draining from the publish queue

func main() {
	shutdown, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)

	broker := pubsub.Broker()
	publisher := pubsub.Publisher("counter", broker)
	sub1 := pubsub.Subscriber("sub1", "counter", broker)
	sub2 := pubsub.Subscriber("sub2", "counter", broker)

	taskManger := goasync.NewTaskManager(goasync.StrictConfig())
	taskManger.AddTask("broker", broker)
	taskManger.AddTask("publisher", publisher)
	taskManger.AddTask("subscriber 1", sub1)
	taskManger.AddTask("subscriber 2", sub2)

	if errs := taskManger.Run(shutdown); errs != nil {
		log.Fatal(errs)
	}

}
