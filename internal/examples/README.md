# Examples

The examples provide some example use cases for running multi threaded applications
and how they can be setup with goasync. Each of the `main` files can be found under
the `cmd/[example]` directories.

Each of the examples can be ran with:
```
# run with race detection enabled
go run --race main.go
```

## pubsub

A very basic feature set for a pubsub sysstem. In this case, the clients/subscribers process messages
slower than the publisher and allow for a buffer of 5 before messages start dropping. However when
Ctrl-C is pressed for the program, the clients will at least drain any availabe messages before shutting down
