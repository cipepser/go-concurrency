package heartbeat

import (
	"testing"
)

func TestDoWorkMock(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	intSlice := []int{0, 1, 2, 3, 5}
	heartbeat, results := DoWorkMock(done, intSlice...)

	<-heartbeat

	for i, expected := range intSlice {
		select {
		case r := <-results:
			if r != expected {
				t.Errorf("index %v: expected %v, but received %v", i, expected, r)
			}
		}
	}

	// bad case: use `case <-time.After(1 * time.Second):`
	//❯ go test ./heartbeat
	//--- FAIL: TestDoWorkMock (1.01s)
	//    heartBeat_test.go:22: test timed out
	//FAIL
	//FAIL	github.com/cipepser/go-concurrency/chap5/heartbeat	1.013s

	// good case: use `<-heartbeat`
	//❯ go test ./heartbeat
	//ok  	github.com/cipepser/go-concurrency/chap5/heartbeat	2.010s
}
