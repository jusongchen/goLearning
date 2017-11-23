package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

//Job describes a job
type Job struct{ jobID int64 }

func process(job Job) {
	fmt.Printf("working on job %v\n", job.jobID)
}

func genJobs(ctx context.Context, jobChan chan<- Job, jobPerSec int, maxJobs int64) {
	// const estimatedRunTime = time.Duration(0)

	tickInterval := time.Duration(int64((time.Second - estimatedRunTime)) / int64(jobPerSec))
	jobID := int64(0)
	for _ = range time.Tick(tickInterval) { //control the pace
		select {
		case <-ctx.Done():
			return //no more requests needed - done signal detected
		default:
			job := Job{jobID: jobID}
			go TryEnqueue(ctx, job, jobChan) //we do not wait to enque to complete
			jobID++
			if jobID == maxJobs {
				return
			}
		}
	}
}

const degree_of_parallel = 1024
const max_jobs_in_queue int = 1e10

const numJobs int64 = 50000
const durationInSec = 5

//TODO:How to estimate this??
const estimatedRunTime = 50 * time.Millisecond

func main() {
	// make a channel with a capacity of 100.

	jobChan := make(chan Job, max_jobs_in_queue)
	ctx, cancel := context.WithTimeout(context.Background(), durationInSec*time.Second)

	go func() {
		processJobs(ctx, jobChan, degree_of_parallel)
	}()
	go func() {

		rate := int(numJobs / durationInSec)
		genJobs(ctx, jobChan, rate, numJobs)
	}()
	startTime := time.Now()
	<-ctx.Done()
	cancel()
	log.Printf("Elapsed time %v ", time.Since(startTime))

}
