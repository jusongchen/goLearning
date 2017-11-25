package main

import (
	"context"
	"sync"
	"time"
)

func worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan Request, resChan chan<- Response, errChan chan<- error) {
	defer wg.Done()
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
			if err != nil {
				errChan <- err
			} else {
				resChan <- res
			}

			//flag is set before a cancel signal is sent out.
			//when a Cancel signal is send out, neither jobChan nor ctx.Done() will be blocked.
			//if the flag is set, we drop the rest and get out of here!
			// if atomic.LoadUint64(&flagQuit) == 1 {
			// 	return
			// }
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
func doJobs(ctx context.Context, workerCount int, rampDownPeriod time.Duration, jobChan <-chan Request, resChan chan<- Response, errChan chan<- error) {

	// use a WaitGroup
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, &wg, jobChan, resChan, errChan)
	}

	//wait at most 1 more second after done signal
	WaitTimeout(ctx, rampDownPeriod, &wg)
	close(resChan)
	close(errChan)
}

func processJobs(ctx context.Context, workerCount int, rampDownPeriod time.Duration, jobChan <-chan Request, resChan chan<- Response, errChan chan<- error) (aborted int64) {
	// var flagQuit uint64

	c, cancel := context.WithCancel(context.Background())
	go doJobs(c, workerCount, rampDownPeriod, jobChan, resChan, errChan)

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
