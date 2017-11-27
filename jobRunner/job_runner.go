package runner

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

//worker process the job request
//After job is done, it sends result to resChan. it will discard results if  distcardResult signal received.
func worker(discardResult <-chan struct{}, wg *sync.WaitGroup, req Request, resChan chan<- Response, errChan chan<- error, handleFunc func(job Request) (Response, error)) {

	defer func() {
		if rvr := recover(); rvr != nil {
			fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
			debug.PrintStack()
		}
		wg.Done()
	}()

	res, err := handleFunc(req)

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

}

// WaitTimeout returns when either 1) wg *sync.WaitGroup is donw 2) graceTime passed after the ctx.done() signal
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

//processJobs:
//  invoke workers at fix rate to do jobs
//	close resChan and errChan when returns
func processJobs(ctx context.Context, rate float64, runPeriod, rampDown time.Duration, resChan chan<- Response, errChan chan<- error, handleFunc func(job Request) (Response, error)) {
	// var flagQuit uint64

	// use a WaitGroup
	var wg sync.WaitGroup

	//to signal if all processed results should be discarded
	discardResult := make(chan struct{})

	throttle := time.Tick(time.Second / time.Duration(rate))

	//run period
	ctxRun, cancel := context.WithTimeout(ctx, runPeriod)
	defer cancel()

	newJob := newReqFunc()

	//calculate max request # at this rate.
	maxReq := int64((float64(runPeriod) / float64(time.Second)) * rate)

ForLoop:
	for i := int64(0); i < maxReq; i++ {
		select {
		case <-ctxRun.Done(): //rampup and steady Period timeout
			break ForLoop
		default:
			req := newJob()
			wg.Add(1)
			go worker(discardResult, &wg, req, resChan, errChan, handleFunc)
			<-throttle
		}
	}

	//WaitTimeout returns when
	//1) all waitGroup are done, or
	//2) rampDown period has passed after ctx.done() signal
	WaitTimeout(ctx, rampDown, &wg)

	//send signal to ask worker to abandon any processed results as we are closing result channels
	close(discardResult)

	close(resChan)
	close(errChan)

}

type stat struct {
	begin     time.Time
	end       time.Time
	okCnt     int64
	errCnt    int64
	minDoTime time.Duration
	maxDoTime time.Duration
	sumDoTime time.Duration
}

//Request describes a job
type Request struct{ jobID int64 }

//Response describes process result
type Response struct {
	elapsed time.Duration
	msg     string
}

func newReqFunc() func() Request {
	jobID := int64(0)
	return func() Request {
		r := Request{
			jobID: jobID,
		}
		jobID++
		return r
	}
}

func processResult(ctx context.Context, resChan <-chan Response, errChan <-chan error) (s stat, eof bool) {

	s = stat{
		begin: time.Now(),
	}
	defer func() {
		s.end = time.Now()
	}()

forLoop:
	for {
		select {
		case <-ctx.Done():
			break forLoop
		default:
			select {
			case res, ok := <-resChan:
				if !ok { //resChan closed
					resChan = nil

				} else {
					s.okCnt++
					s.sumDoTime += res.elapsed
					if res.elapsed > s.maxDoTime {
						s.maxDoTime = res.elapsed
					}
					if res.elapsed < s.minDoTime {
						s.minDoTime = res.elapsed
					}
				}

			case err, ok := <-errChan:
				if !ok { //errChan closed
					errChan = nil
				} else {
					s.errCnt++
					_ = err
				}
			default:
				if errChan == nil && resChan == nil {
					eof = true
					break forLoop
				}
			}
		}
	}
	return s, eof
}

//See https://en.wikipedia.org/wiki/Token_bucket
func run(runPeriod, rampDown time.Duration, rate float64, burst int, statInterval time.Duration, statChan chan<- stat, handleFunc func(job Request) (Response, error)) {

	resChan := make(chan Response, burst)
	errChan := make(chan error, burst)

	totalDuration := runPeriod + rampDown
	ctx, cancel := context.WithTimeout(context.Background(), totalDuration)
	defer cancel()

	defer close(statChan)

	go processJobs(ctx, rate, runPeriod, rampDown, resChan, errChan, handleFunc)

	for {
		ctxStat, cancelStat := context.WithTimeout(context.Background(), statInterval)
		s, eof := processResult(ctxStat, resChan, errChan)
		cancelStat()
		statChan <- s
		if eof { //both errChan and resChan closed
			break
		}
	}
}
