package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func replicatedWork() {
	doWork := func(
		done <-chan interface{},
		id int,
		wg *sync.WaitGroup,
		result chan<- int,
	) {
		started := time.Now()
		defer wg.Done()

		// Simulate randome load
		simulatedLoadTime := time.Duration(1+rand.Intn(5)) * time.Second
		select {
		case <-done:
		case result <- id:
		}

		took := time.Since(started)
		// Display how long handlers would have taken
		if took < simulatedLoadTime {
			took = simulatedLoadTime
		}
		fmt.Printf("%v took %v\n", id, took)
	}

	done := make(chan interface{})
	result := make(chan int)
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go doWork(done, i, &wg, result)
	}

	firstReturned := <-result
	close(done)
	wg.Wait()

	fmt.Printf("Received an answer from #%v\n", firstReturned)
	//9 took 3s
	//0 took 2s
	//6 took 1s
	//8 took 1s
	//7 took 1s
	//5 took 5s
	//4 took 2s
	//1 took 3s
	//2 took 4s
	//3 took 2s
	//Received an answer from #9
}

func main() {
	replicatedWork()
}
