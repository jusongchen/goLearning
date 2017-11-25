package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

//Request describes a job
type Request struct{ jobID int64 }

//Response describes process result
type Response struct {
	msg string
}

func process(job Request) (Response, error) {
	res := Response{
		msg: fmt.Sprintf("processing job %v", job.jobID),
	}
	return res, nil

}

func main() {
	// make a channel with a capacity of 100.

	const degree_of_parallel = 1024
	const max_jobs_in_queue int = degree_of_parallel * 8

	const numJobs int64 = 50000
	const TotalDuration = 5 * time.Second

	//TODO:How to estimate this??
	const rampDownPeriod = 50 * time.Millisecond

	jobChan := make(chan Request, max_jobs_in_queue)
	resChan := make(chan Response, max_jobs_in_queue)
	errChan := make(chan error, max_jobs_in_queue)
	ctx, cancel := context.WithTimeout(context.Background(), TotalDuration)

	aborted := int64(0)
	doneOk := int64(0)
	errOut := int64(0)

	go func() {
		aborted = processJobs(ctx, degree_of_parallel, rampDownPeriod, jobChan, resChan, errChan)
	}()
	go func() {
		loadJobs(ctx, TotalDuration, rampDownPeriod, numJobs, jobChan)
	}()

	go func() {
		for res := range resChan {
			doneOk++
			fmt.Printf("response:%v\n", res.msg)
		}
	}()

	go func() {
		for err := range errChan {
			errOut++
			fmt.Printf("error:%v\n", err)
		}
	}()

	startTime := time.Now()
	<-ctx.Done()
	cancel()
	log.Printf("Elapsed time %v , done %v,errors %v, aborted %v", time.Since(startTime), doneOk, errOut, aborted)

}
