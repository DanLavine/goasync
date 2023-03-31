package pubsub

import (
	"context"
	"fmt"
	"time"
)

type subscriber struct {
	name    string
	channel string
	broker  *broker

	messageQueue <-chan interface{}
}

func Subscriber(name string, channel string, broker *broker) *subscriber {
	return &subscriber{
		name:    name,
		channel: channel,
		broker:  broker,
	}
}

func (s *subscriber) Initialize() error {
	// subscribe to the message broker
	s.messageQueue = s.broker.Subscribe(s.channel, 5)
	return nil
}

func (s *subscriber) Cleanup() error {
	return nil
}

func (s *subscriber) Execute(ctx context.Context) error {
	for {
		select {
		case message, ok := <-s.messageQueue:

			if !ok {
				select {
				case <-ctx.Done():
					// initiated a proper shutdown
					return nil
				default:
					// something failed and closed our connection
					return fmt.Errorf("message client closed unexpectedly")
				}
			}

			fmt.Printf("%s received message: %v\n", s.name, message)

			// sleep longer than the publisher to ensure we build a buffer and drain everything!
			time.Sleep(2 * time.Second)
		}
	}
}
