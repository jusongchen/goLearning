package main

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func jobHandleFunc(errRate float64, jobTime time.Duration) func(job Request) (Response, error) {
	rand.Seed(time.Now().UTC().UnixNano())

	return func(job Request) (Response, error) {
		time.Sleep(jobTime)
		var err error

		res := Response{}

		f := rand.Float64()
		if f < errRate {
			err = fmt.Errorf("handling job %v \tError out", job.jobID)
			fmt.Printf("%v\r", err)

		} else {
			res.msg = fmt.Sprintf("handling job %v\tOK", job.jobID)
			fmt.Printf("%v\r", res)
		}
		return res, err
	}
}

type testCase struct {
	name            string
	burst           int
	rate            float64
	runPeriod       time.Duration
	rampDown        time.Duration
	errRate         float64
	jobTime         time.Duration
	expectedRunTime time.Duration
	expectedOkCnt   int64
	expectedErrCnt  int64
}

var rates = []float64{1, 20, 300, 4000, 5000}
var bursts = []int{1, 4096}
var runTimes = []time.Duration{2 * time.Second, 10 * time.Second, 60 * time.Second}
var errorRates = []float64{0, 0.1}
var jobMaxExecTimes = []time.Duration{0, 1 * time.Millisecond, 10 * time.Millisecond, 100 * time.Millisecond, 1 * time.Second}

func TestShort(t *testing.T) {
	rates = []float64{1, 20}
	bursts = []int{1, 4096}
	runTimes = []time.Duration{2 * time.Second}
	errorRates = []float64{0, 0.1}
	jobMaxExecTimes = []time.Duration{1 * time.Millisecond, 10 * time.Millisecond}
	runTests(t)
}

func TestLong(t *testing.T) {

	rates = []float64{1, 20, 300, 4000, 5000}
	bursts = []int{1, 4096}
	runTimes = []time.Duration{2 * time.Second, 10 * time.Second, 60 * time.Second}
	errorRates = []float64{0, 0.1}
	jobMaxExecTimes = []time.Duration{0, 1 * time.Millisecond, 10 * time.Millisecond, 100 * time.Millisecond, 1 * time.Second}

	runTests(t)
}

func runTests(t *testing.T) {

	rampDown := 500 * time.Millisecond

	tt := []testCase{}

	for _, d := range runTimes {
		for _, burst := range bursts {
			for _, errRate := range errorRates {
				for _, rate := range rates {
					for _, jobTime := range jobMaxExecTimes {
						tc :=
							testCase{
								name:            fmt.Sprintf("rate_%v,\tburst_%v,\trunTime %v,\terrRate %v,\tjobTime %v", rate, burst, d, errRate, jobTime),
								burst:           burst,
								rate:            rate,
								runPeriod:       d,
								rampDown:        rampDown,
								errRate:         errRate,
								jobTime:         jobTime,
								expectedRunTime: d,
								expectedOkCnt:   int64((float64(d / time.Second)) * rate * (1.0 - errRate)),
								expectedErrCnt:  int64((float64(d / time.Second)) * rate * errRate),
							}
						tt = append(tt, tc)
					}
				}
			}
		}
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {

			startTime := time.Now()
			okCnt, errCnt := runner(tc.runPeriod, tc.rampDown, tc.rate, tc.burst, jobHandleFunc(tc.errRate, tc.jobTime))
			endTime := time.Now()
			elapsed := time.Since(startTime)
			log.Printf("case %v\t:Elapsed time %v,\tdone %v,\terrors %v", tc, elapsed, okCnt, errCnt)
			assert.WithinDuration(t, startTime.Add(tc.expectedRunTime), endTime, tc.rampDown)

			if tc.jobTime > tc.rampDown {
				return
			}

			//when jobTime is greater than ramp downTime, job results discarded
			//we expected all job to be handled unless
			if okCnt != 0 && tc.expectedOkCnt > 10 {
				assert.InEpsilon(t, tc.expectedOkCnt, okCnt, 0.2, "OKcnt expected %v actual %v", tc.expectedOkCnt, okCnt)
			}
			if errCnt != 0 && tc.expectedErrCnt > 10 {
				assert.InEpsilon(t, tc.expectedErrCnt, errCnt, 0.2, "ErrCnt expected %v actual %v", tc.expectedErrCnt, errCnt)
			}
		})
	}

}
