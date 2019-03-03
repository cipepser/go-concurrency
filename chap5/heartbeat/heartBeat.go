package heartbeat

import (
	"fmt"
	"math/rand"
	"time"
)

func DoWork() {
	doWork := func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (<-chan interface{}, <-chan time.Time) {
		heartbeat := make(chan interface{})
		results := make(chan time.Time)
		go func() {
			//defer close(heartbeat)
			//defer close(results)

			pulse := time.Tick(pulseInterval)
			workGen := time.Tick(2 * pulseInterval)

			sendPulse := func() {
				select {
				case heartbeat <- struct{}{}:
				default:
				}
			}

			sendResult := func(r time.Time) {
				for {
					select {
					case <-done:
						return
					case <-pulse:
						sendPulse()
					case results <- r:
						return
					}
				}
			}

			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case r := <-workGen:
					sendResult(r)
				}
			}
		}()
		return heartbeat, results
	}

	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() {
		close(done)
	})

	const timeout = 2 * time.Second
	heartbeat, results := doWork(done, timeout/2)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Printf("results %v\n", r.Second())
		case <-time.After(timeout):
			fmt.Println("worker goroutine is not healthy!")
			return
		}
	}
	// correct pattern
	//pulse
	//pulse
	//results 25
	//pulse
	//pulse
	//results 27
	//pulse
	//pulse
	//results 29
	//pulse
	//pulse
	//results 31
	//pulse
	//pulse
	//results 33

	// bad pattern(doesn't close heartbeat and results channels
	//pulse
	//pulse
	//worker goroutine is not healthy!
}

func heartbeatForEachWork() {
	doWork := func(
		done <-chan interface{},
	) (<-chan interface{}, <-chan int) {
		heartbeatStream := make(chan interface{}, 1)
		workStream := make(chan int)
		go func() {
			defer close(heartbeatStream)
			defer close(workStream)

			for i := 0; i < 10; i++ {
				select {
				case heartbeatStream <- struct{}{}:
				default:
				}

				select {
				case <-done:
					return
				case workStream <- rand.Intn(10):
				}
			}
		}()

		return heartbeatStream, workStream
	}

	done := make(chan interface{})
	defer close(done)

	heartbeat, results := doWork(done)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Printf("results %v\n", r)
		}
	}

	// 1st execution
	//pulse
	//results 1
	//pulse
	//results 7
	//pulse
	//results 7
	//pulse
	//results 9
	//pulse
	//results 1
	//pulse
	//results 8
	//pulse
	//results 5
	//pulse
	//results 0
	//pulse
	//results 6
	//pulse
	//results 0

	// 2nd execution
	//pulse
	//results 1
	//pulse
	//results 7
	//pulse
	//results 7
	//results 9
	//pulse
	//pulse
	//results 1
	//pulse
	//results 8
	//pulse
	//results 5
	//pulse
	//results 0
	//results 6
	//pulse
	//pulse
	//results 0
}

func DoWorkMock(done <-chan interface{}, nums ...int) (<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1)
	intStream := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(intStream)

		time.Sleep(2 * time.Second) // simulate to do something

		for _, n := range nums {
			select {
			case heartbeat <- struct{}{}:
			default:
			}

			select {
			case <-done:
				return
			case intStream <- n:
			}
		}
	}()

	return heartbeat, intStream
}

//func main() {
//	//DoWork()
//
//	heartbeatForEachWork()
//}
