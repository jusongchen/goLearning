package main

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
)

var flag uint64

func worker(ctx context.Context, jobChan <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return

		case job := <-jobChan:
			process(job)
			if atomic.LoadUint64(&flag) == 1 {
				return
			}
		}
	}
}

// WaitWithContext does a Wait on a sync.WaitGroup object but with a specified
// timeout. Returns true if the wait completed without timing out, false
// otherwise.
func WaitWithContext(ctx context.Context, wg *sync.WaitGroup) bool {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		return true
	case <-ctx.Done():
		return false
	}
}

func doJobs(ctx context.Context, jobChan <-chan Job, workerCount int) {

	// use a WaitGroup
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, jobChan)
	}

	// now use the WaitTimeout instead of wg.Wait()
	WaitWithContext(ctx, &wg)

}

func processJobs(ctx context.Context, jobChan <-chan Job, workerCount int) {

	c, cancel := context.WithCancel(context.Background())
	go doJobs(c, jobChan, workerCount)

	//wait for done signal
	<-ctx.Done()

	// set the flag first, before cancelling
	atomic.StoreUint64(&flag, 1)
	cancel()

}

// TryEnqueue tries to enqueue a job to the given job channel. Returns true if
// the operation was successful, and false if enqueuing would not have been
// possible without blocking. Job is not enqueued in the latter case.
func TryEnqueue(ctx context.Context, job Job, jobChan chan<- Job) bool {

	for {
		select {
		case <-ctx.Done():
			log.Printf("Timeout:Job not enqued")
			return false
		case jobChan <- job:
			return true
		default:
			//will retry enque
		}
	}
}
