package runner

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
		begin := time.Now()
		res := Response{}
		defer func() {
			res.elapsed = time.Since(begin)
		}()
		time.Sleep(jobTime)
		var err error

		f := rand.Float64()
		if f < errRate {
			err = fmt.Errorf("handling job %v \tError out", job.jobID)
			// fmt.Printf("%v\r", err)

		} else {
			res.msg = fmt.Sprintf("handling job %v\tOK", job.jobID)
			// fmt.Printf("%v\r", res)
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
	expectedOkRate  float64
	expectedErrRate float64
}

func TestShort(t *testing.T) {
	rates := []float64{1, 20}
	bursts := []int{1, 4096}
	runTimes := []time.Duration{2 * time.Second, 10 * time.Second}
	errorRates := []float64{0, 0.2}
	jobMaxExecTimes := []time.Duration{1 * time.Millisecond, 1 * time.Second}
	runTests(t, rates, bursts, runTimes, errorRates, jobMaxExecTimes)
}

func TestLong(t *testing.T) {

	rates := []float64{1, 20, 300, 4000, 10000}
	bursts := []int{1, 4096}
	runTimes := []time.Duration{2 * time.Second, 10 * time.Second, 60 * time.Second}
	errorRates := []float64{0, 0.3}
	jobMaxExecTimes := []time.Duration{0, 10 * time.Millisecond, 100 * time.Millisecond, 1 * time.Second}

	runTests(t, rates, bursts, runTimes, errorRates, jobMaxExecTimes)
}

func runTests(t *testing.T, rates []float64, bursts []int, runTimes []time.Duration, errorRates []float64, jobMaxExecTimes []time.Duration) {

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
								expectedOkRate:  rate * (1.0 - errRate),
								expectedErrRate: rate * errRate,
							}
						tt = append(tt, tc)
					}
				}
			}
		}
	}

	for _, tc := range tt {

		t.Run(tc.name, func(t *testing.T) {
			statInterval := 2 * time.Second
			statChan := make(chan stat)
			go run(tc.runPeriod, tc.rampDown, tc.rate, tc.burst, statInterval, statChan, jobHandleFunc(tc.errRate, tc.jobTime))

			okCnt, errCnt := int64(0), int64(0)
			startTime := time.Now()
			for s := range statChan {
				okCnt += s.okCnt
				errCnt += s.errCnt
				interval := s.end.Sub(s.begin)
				okRate := s.okCnt * int64(time.Second) / int64(interval)
				errRate := s.errCnt * int64(time.Second) / int64(interval)

				if okRate != 0 && tc.expectedOkRate > 0 {
					assert.InEpsilon(t, tc.expectedOkRate, okRate, 0.5, "%+v %+v okRate expected %v actual %v", tc, s, tc.expectedOkRate, okRate)
					// log.Printf("%+v %+v okRate expected %v actual %v", tc, s, tc.expectedOkRate, okRate)
				}
				if errRate != 0 && tc.expectedErrRate > 0 {
					assert.InEpsilon(t, tc.expectedErrRate, errRate, 0.5, "%+v %+v errRate expected %v actual %v", tc, s, tc.expectedErrRate, errRate)
					// log.Printf("%+v %+v errRate expected %v actual %v", tc, s, tc.expectedErrRate, errRate)
				}

				avgDoTime := time.Duration(0)
				if s.okCnt != 0 {
					avgDoTime = s.sumDoTime / time.Duration(s.okCnt)
				}
				if avgDoTime != 0 {
					assert.InEpsilon(t, tc.jobTime, avgDoTime, 0.5, "%+v %+v avgDoTime expected %v actual %v", tc, s, tc.jobTime, avgDoTime)
				}
			}
			elapsed := time.Since(startTime)

			log.Printf("case %+v\n:Elapsed time %v,\tdone %v,\terrors %v", tc, elapsed, okCnt, errCnt)
			delta := tc.rampDown
			if tc.jobTime > delta {
				delta = tc.jobTime
			}
			assert.InDelta(t, int64(tc.expectedRunTime), int64(elapsed), float64(delta+500*time.Millisecond), "case %+v\t:Elapsed time %v,\tdone %v,\terrors %v", tc, elapsed, okCnt, errCnt)

			if tc.jobTime > tc.rampDown {
				return
			}

			//when jobTime is greater than ramp downTime, job results discarded
			//we expected all job to be handled unless
			if okCnt != 0 && tc.expectedOkCnt > 10 {
				assert.InEpsilon(t, tc.expectedOkCnt, okCnt, 0.2, "case %+v OKcnt expected %v actual %v", tc, tc.expectedOkCnt, okCnt)
			}
			if errCnt != 0 && tc.expectedErrCnt > 10 {
				assert.InEpsilon(t, tc.expectedErrCnt, errCnt, 0.2, "case %+v ErrCnt expected %v actual %v", tc, tc.expectedErrCnt, errCnt)
			}
		})
	}

}
