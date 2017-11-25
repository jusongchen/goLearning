package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

//worker takes two context, one to stop take in any further work, another to discard any processed results
func worker(ctx context.Context, discardResult <-chan struct{}, wg *sync.WaitGroup, jobChan <-chan Request, resChan chan<- Response, errChan chan<- error) {

	defer func() {
		if rvr := recover(); rvr != nil {
			fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
			debug.PrintStack()
		}
		wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobChan:
			if !ok {
				//jobChan closed
				return
			}
			res, err := process(job)

			select {

			case <-discardResult: //controller ask to discardResult as out channels have been closed
				fmt.Fprintf(os.Stderr, "Result discard: %+v, err: %v\n", res, err)
				return
			default:
				if err != nil {
					//if errChan closed , deferred recover() will save us
					errChan <- err
				} else {
					//if resChan closed , deferred recover() will save us
					resChan <- res
				}
			}

			select {
			//check to see if done signal is sent
			case <-ctx.Done():
				return
			default:
			}
		}
	}
}

// WaitTimeout does a Wait on a sync.WaitGroup object but with a specified
// timeout after the ctx is done.
func WaitTimeout(ctx context.Context, graceTime time.Duration, wg *sync.WaitGroup) {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		return
	case <-ctx.Done():
		//cancel signal detected, but the waitgroup not done yet. wait a graceful period
		<-time.After(graceTime)
		return
	}
}

//doJobs process jobs concurrently
//close resChan and errChan when return
func jobCoordinator(ctx context.Context, workerCount int, rampDownPeriod time.Duration, jobChan <-chan Request, resChan chan<- Response, errChan chan<- error) {

	// use a WaitGroup
	var wg sync.WaitGroup

	//to signal if all processed results should be discarded
	discardResult := make(chan struct{})

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, discardResult, &wg, jobChan, resChan, errChan)
	}

	//WaitTimeout returns when
	//1) all waitGroup are done, or
	//2) rampDownPeriod has passed
	WaitTimeout(ctx, rampDownPeriod, &wg)

	//send signal to ask worker to abandon any processed results as we are closing result channels
	close(discardResult)

	close(resChan)
	close(errChan)
}

func processJobs(ctx context.Context, workerCount int, rampDownPeriod time.Duration, jobChan <-chan Request, resChan chan<- Response, errChan chan<- error) (aborted int64) {
	// var flagQuit uint64

	c, cancel := context.WithCancel(context.Background())
	go jobCoordinator(c, workerCount, rampDownPeriod, jobChan, resChan, errChan)

	//wait for done signal
	<-ctx.Done()

	// set the cancel flag first, before cancelling
	// atomic.StoreUint64(&flagQuit, 1)

	aborted = int64(0)
	//now start drain jobChan
	for _ = range jobChan {
		aborted++
	}

	cancel() //send done() signal to all workers
	return aborted

}

//loadJobs send jobs to queue
//return when ctx.Done() and close jobChan
func loadJobs(ctx context.Context, duration time.Duration, rampDownPeriod time.Duration, numJobs int64, jobChan chan<- Request) int64 {

	defer close(jobChan)
	tickInterval := time.Duration(int64(duration-rampDownPeriod) / numJobs)
	jobSent := int64(0)
	for _ = range time.Tick(tickInterval) { //control the pace
		select {
		case <-ctx.Done():
			return jobSent //no more requests needed - done signal detected
		default:
			job := Request{jobID: jobSent}

			select {
			case jobChan <- job:
				// job enqueued successfully
			default:
				// cannot enqueue immediately. retry enque but we do not know how long it may take
				//so do it async
				go func() {
					jobChan <- job
					//we will drain the jobChan after timeout, hence this send won't be blocked forever
				}()
			}
			jobSent++
			if jobSent == numJobs {
				return jobSent
			}
		}
	}
	return jobSent
}
