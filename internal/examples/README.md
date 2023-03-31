# Examples

The examples provide some example use cases for running multi threaded applications
and how they can be setup with goasync. Each of the `main` files can be found under
the `cmd/[example]` directories. Each application can be stopped with Ctrl-c

Each of the examples can be ran with:
```
# run with race detection enabled
go run --race main.go
```

## force-shutdown

A basic example of what can happen when one sub task goes rouge and doesn't respect shutdown signals.
For this test case, there is a 5 second timeout set on the Force Stop Task which ensures that the
program will exit.

## pubsub

A very basic feature set for a pubsub sysstem. In this case, the clients/subscribers process messages
slower than the publisher and allow for a buffer of 5 before messages start dropping. However when
Ctrl-c is pressed for the program, the clients will at least drain any availabe messages before shutting down
