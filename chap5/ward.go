package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		orDone := make(chan interface{})
		go func() {
			defer close(orDone)
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}

	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if !ok {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}

				}
			}
		}()
		return valStream
	}

	bridge := func(
		done <-chan interface{},
		chanStream <-chan <-chan interface{},
	) <-chan interface{} {
		valSteam := make(chan interface{})
		go func() {
			defer close(valSteam)
			for {
				var stream <-chan interface{}
				select {
				case maybeStream, ok := <-chanStream:
					if !ok {
						return
					}
					stream = maybeStream
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					select {
					case valSteam <- val:
					case <-done:
					}
				}
			}
		}()
		return valSteam
	}

	take := func(
		done <-chan interface{},
		valueStream <-chan interface{},
		num int,
	) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}
		}()
		return takeStream
	}

	type startGoroutineFn func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (heartbeat <-chan interface{})

	newSteward := func(
		timeout time.Duration,
		startGoroutine startGoroutineFn,
	) startGoroutineFn {
		return func(
			done <-chan interface{},
			pulseInterval time.Duration,
		) <-chan interface{} {
			heartbeat := make(chan interface{})
			go func() {
				defer close(heartbeat)

				var wardDone chan interface{}
				var wardHeartbeat <-chan interface{}
				startWard := func() {
					wardDone = make(chan interface{})
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
				}
				startWard()
				pulse := time.Tick(pulseInterval)

			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case <-wardHeartbeat:
							continue monitorLoop
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy; restarting")
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}
				}
			}()

			return heartbeat
		}
	}

	doWorkFn := func(
		done <-chan interface{},
		intList ...int,
	) (startGoroutineFn, <-chan interface{}) {
		intChanStream := make(chan (<-chan interface{}))
		intStream := bridge(done, intChanStream)
		doWork := func(
			done <-chan interface{},
			pulseInterval time.Duration,
		) <-chan interface{} {
			intStream := make(chan interface{})
			heartbeat := make(chan interface{})
			go func() {
				defer close(intStream)
				select {
				case intChanStream <- intStream:
				case <-done:
					return
				}

				pulse := time.Tick(pulseInterval)

				for {
				valueLoop:
					for _, intVal := range intList {
						if intVal < 0 {
							log.Printf("negative value: %v\n", intVal)
							return
						}

						for {
							select {
							case <-pulse:
								select {
								case heartbeat <- struct{}{}:
								default:
								}
							case intStream <- intVal:
								continue valueLoop
							case <-done:
								return
							}
						}
					}
				}
			}()
			return heartbeat
		}
		return doWork, intStream
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	//doWork := func(done <-chan interface{}, _ time.Duration) <-chan interface{} {
	//	log.Println("ward: Hello, I'm irresponsible!")
	//	go func() {
	//		<-done
	//		log.Println("ward: I am halting.")
	//	}()
	//	return nil
	//}
	//doWorkWithSteward := newSteward(4*time.Second, doWork)
	//
	//done := make(chan interface{})
	//time.AfterFunc(9*time.Second, func() {
	//	log.Println("main: halting steward and wart.")
	//	close(done)
	//})
	//
	//for range doWorkWithSteward(done, 4*time.Second) {
	//}
	//log.Println("Done")

	//08:50:14 ward: Hello, I'm irresponsible!
	//08:50:18 steward: ward unhealthy; restarting
	//08:50:18 ward: Hello, I'm irresponsible!
	//08:50:18 ward: I am halting.
	//08:50:22 steward: ward unhealthy; restarting
	//08:50:22 ward: Hello, I'm irresponsible!
	//08:50:22 ward: I am halting.
	//08:50:23 main: halting steward and wart.
	//08:50:23 Done
	//08:50:23 ward: I am halting.

	done := make(chan interface{})
	defer close(done)

	doWork, intStream := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range take(done, intStream, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}
	//09:10:19 negative value: -1
	//Received: 1
	//Received: 2
	//09:10:19 steward: ward unhealthy; restarting
	//09:10:19 negative value: -1
	//Received: 1
	//Received: 2
	//09:10:19 steward: ward unhealthy; restarting
	//Received: 1
	//Received: 2
	//09:10:19 negative value: -1
}
