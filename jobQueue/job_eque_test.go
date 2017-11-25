package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {

	tt := []struct {
		name            string
		DOP             int
		queueSizeFactor int
		numJobs         int64
		totalDuration   time.Duration
		rampDownPeriod  time.Duration
		expectedRunTime time.Duration
		expectedAborted int64
		expectedDoneOK  int64
		expectedErr     int64
	}{
		{"DOP 1", 1, 1, 99, 5 * time.Second, 50 * time.Millisecond, 5 * time.Second, 0, 99, 0},
		{"DOP 10", 10, 1, 99, 5 * time.Second, 50 * time.Millisecond, 5 * time.Second, 0, 99, 0},
		{"DOP 1024", 1024, 1, 99, 5 * time.Second, 50 * time.Millisecond, 5 * time.Second, 0, 99, 0},
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {

			queueSize := tc.DOP * tc.queueSizeFactor
			jobChan := make(chan Request, queueSize)
			resChan := make(chan Response, queueSize)
			errChan := make(chan error, queueSize)

			ctx, cancel := context.WithTimeout(context.Background(), tc.totalDuration)

			aborted := int64(0)
			doneOk := int64(0)
			errOut := int64(0)
			jobSent := int64(0)

			go func() {
				aborted = processJobs(ctx, tc.DOP, tc.rampDownPeriod, jobChan, resChan, errChan)
			}()
			go func() {
				jobSent = loadJobs(ctx, tc.totalDuration, tc.rampDownPeriod, tc.numJobs, jobChan)
			}()

			go func() {
				for res := range resChan {
					doneOk++
					// fmt.Printf("response:%v\n", res.msg)
					_ = res
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
			elapsed := time.Since(startTime)
			log.Printf("test case %v:Elapsed time %v , done %v,errors %v, aborted %v", tc.name, elapsed, doneOk, errOut, aborted)
			assert.InEpsilon(t, tc.expectedRunTime, elapsed, 0.01, "runtime ")
			assert.Equal(t, tc.expectedDoneOK, doneOk)
			assert.Equal(t, tc.expectedAborted, aborted)
			assert.Equal(t, tc.expectedErr, errOut)

		})
	}

}
