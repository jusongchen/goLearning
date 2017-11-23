package main

import (
	"context"
	"fmt"
	"time"
)

//Job describes a job
type Job struct{}

func process(job Job) {
	fmt.Printf("working on job ")
}

func gen_job(ctx context.Context, jobChan chan<- Job, jobPerSec int) {

	tickInterval := time.Nanosecond * time.Duration(1e9/(jobPerSec))

	for _ = range time.Tick(tickInterval) { //control the pace
		select {
		case <-ctx.Done():
			return //no more requests needed - done signal detected
		default:
			job := Job{}
			go TryEnqueue(ctx, job, jobChan) //we do not wait to enque to complete
		}
	}
}

const degree_of_parallel = 32
const max_jobs_in_queue int = 1e10

func Test() {
	// make a channel with a capacity of 100.
	jobChan := make(chan Job, max_jobs_in_queue)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	processJobs(ctx, jobChan, degree_of_parallel)
	cancel()

}
