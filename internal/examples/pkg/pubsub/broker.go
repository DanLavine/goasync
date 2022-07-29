package pubsub

import (
	"sync"

	"golang.org/x/net/context"
)

// channels are setup to store the
// map[read only channel sent to clients]write only channel side used by broker
type clients map[<-chan interface{}]chan<- interface{}

type broker struct {
	subscriberLock *sync.Mutex

	// save off both a reader and writer version of a client. The Broker needs access to the write client
	// and the subscriber only needs access to a reader client. We don't want a user to close a client,
	// they should Unsubscribe(...) if they don't need to listen anymore
	//
	// map[channel name]
	channels map[string]clients
}

func Broker() *broker {
	return &broker{
		subscriberLock: &sync.Mutex{},
		channels:       map[string]clients{},
	}
}

// Subscribes creates a message client for a particular channel.
//
// Args:
//  - channel: name of the channel to subsribe to
//  - buffer: buffer size for the client messages.
func (b *broker) Subscribe(channel string, buffer int) <-chan interface{} {
	b.subscriberLock.Lock()
	defer b.subscriberLock.Unlock()

	client := make(chan interface{}, buffer)

	// create the named channel if it does not exist
	if _, ok := b.channels[channel]; !ok {
		b.channels[channel] = clients{client: client}
	} else {
		b.channels[channel][client] = client
	}

	return client
}

// Publish sends a message to any number of active clients
func (b *broker) Publish(channel string, msg interface{}) {
	// use an actual lock here, not just a ReadLock. This ensures that we send the message to all our clients
	b.subscriberLock.Lock()
	defer b.subscriberLock.Unlock()

	for _, client := range b.channels[channel] {
		select {
		case client <- msg:
		default:
			// if the client is slow at reading, messages will be dropped
		}
	}
}

// Task Methods
func (b *broker) Initialize() error { return nil }
func (b *broker) Cleanup() error    { return nil }

func (b *broker) Execute(ctx context.Context) error {
	<-ctx.Done()

	b.subscriberLock.Lock()
	defer b.subscriberLock.Unlock()

	// on shutdown, close all clients so they can stop processing requests
	for _, clients := range b.channels {
		// step 1. Close all clients
		for _, writeClient := range clients {
			close(writeClient)
		}
	}

	// step 2. Reset the map to be empty
	b.channels = map[string]clients{}

	return nil
}
