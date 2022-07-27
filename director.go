package goasync

import (
	"context"
)

type director struct {
	namedWorkers []namedWorker
}

func NewDirector() *director {
	return &director{}
}

// Add a worker to the Director
func (d *director) AddWorker(name string, worker Worker) {
	d.namedWorkers = append(d.namedWorkers, namedWorker{name: name, worker: worker})
}

// Run any workers added to the Director.
//
// Rules for Workers:
// 1. Run each Initialize process serially in the order they were added to the Director
//   a. If an error occurs, stop Initializng any remaning workers. Also Run Cleanup for
//      any already workers that have been Initialized
// 2. In Parallel Run all Work(...) functions for any workers
//   a. All workers are expected to run and not error.
//   b. If any workers return an error, the Director will cancel all running workers and then
//      run the Cleanup for each worker.
// 3. Once Stop is called for the Director each worker process will have their context canceled
// 4. Each Worker's Cleanup function is called in reverse order they were added to the Director
func (d *director) Run(ctx context.Context) []NamedError {
	var errors []NamedError

	// initialize
	for index, namedWorker := range d.namedWorkers {
		if err := namedWorker.worker.Initialize(); err != nil {
			errors = append(errors, NamedError{WorkerName: namedWorker.name, Stage: Initialize, Err: err})

			// we hit an error. Run Cleanup in reverse order
			for i := index; i >= 0; i-- {
				if err = d.namedWorkers[i].worker.Cleanup(); err != nil {
					errors = append(errors, NamedError{WorkerName: d.namedWorkers[i].name, Stage: Cleanup, Err: err})
				}
			}

			return errors
		}
	}

	// workers
	//TODO

	// cleanup
	for i := len(d.namedWorkers) - 1; i >= 0; i-- {
		if err := d.namedWorkers[i].worker.Cleanup(); err != nil {
			errors = append(errors, NamedError{WorkerName: d.namedWorkers[i].name, Stage: Cleanup, Err: err})
		}
	}

	return errors
}
