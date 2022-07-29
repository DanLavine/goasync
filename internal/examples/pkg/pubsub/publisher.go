package pubsub

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/net/context"
)

type publisher struct {
	channel string
	broker  *broker
}

func Publisher(channel string, broker *broker) *publisher {
	return &publisher{
		channel: channel,
		broker:  broker,
	}
}

func (p *publisher) Initialize() error {
	rand.Seed(time.Now().UnixNano())
	return nil
}

func (p *publisher) Cleanup() error {
	return nil
}

func (p *publisher) Execute(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	counter := 1
	for {
		select {
		case <-ticker.C:
			fmt.Printf("Sending counter: %d\n", counter)

			p.broker.Publish(p.channel, counter)
			counter++
		case <-ctx.Done():
			return nil
		}
	}
}
