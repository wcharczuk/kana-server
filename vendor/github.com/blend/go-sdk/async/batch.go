/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package async

import (
	"context"
	"runtime"
)

// NewBatch creates a new batch processor.
// Batch processes are a known quantity of work that needs to be processed in parallel.
func NewBatch(work chan interface{}, action WorkAction, options ...BatchOption) *Batch {
	b := Batch{
		Action:      action,
		Work:        work,
		Parallelism: runtime.NumCPU(),
	}
	for _, option := range options {
		option(&b)
	}
	return &b
}

// BatchOption is an option for the batch worker.
type BatchOption func(*Batch)

// OptBatchErrors sets the batch worker error return channel.
func OptBatchErrors(errors chan error) BatchOption {
	return func(i *Batch) {
		i.Errors = errors
	}
}

// OptBatchParallelism sets the batch worker parallelism, or the number of workers to create.
func OptBatchParallelism(parallelism int) BatchOption {
	return func(i *Batch) {
		i.Parallelism = parallelism
	}
}

// Batch is a batch of work executed by a fixed count of workers.
type Batch struct {
	Action      WorkAction
	Parallelism int
	Work        chan interface{}
	Errors      chan error
}

// Process executes the action for all the work items.
func (b *Batch) Process(ctx context.Context) {
	allWorkers := make([]*Worker, b.Parallelism)
	availableWorkers := make(chan *Worker, b.Parallelism)

	// return worker is a local finalizer
	// that grabs a reference to the workers set.
	returnWorker := func(ctx context.Context, worker *Worker) error {
		availableWorkers <- worker
		return nil
	}

	// create and start workers.
	for x := 0; x < b.Parallelism; x++ {
		worker := NewWorker(b.Action)
		worker.Context = ctx
		worker.Errors = b.Errors
		worker.Finalizer = returnWorker

		workerStarted := worker.NotifyStarted()
		go func() { _ = worker.Start() }()
		<-workerStarted

		allWorkers[x] = worker
		availableWorkers <- worker
	}
	defer func() {
		for x := 0; x < len(allWorkers); x++ {
			_ = allWorkers[x].Stop()
		}
	}()

	numWorkItems := len(b.Work)
	var worker *Worker
	var workItem interface{}
	for x := 0; x < numWorkItems; x++ {
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case workItem = <-b.Work:
			select {
			case worker = <-availableWorkers:
				worker.Enqueue(workItem)
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		}
	}

}
