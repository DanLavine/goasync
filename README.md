# GoAsync - Flexible processes management library

Go Async is designed to simplify managing multi-process applications into easily understandable
single purpose units of work.

# Installation

```
go get -u github.com/DanLavine/goasync
```

# Examples

For a few simple and basic real world example check out the `internal/examples` dir

# Running tests

To run all tests:
```
go test --race ./...
```

If we want to run one specific test we can use:
```
go test --race -run [NAME OF TEST] ./...
```
