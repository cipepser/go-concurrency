package main

import (
	"fmt"
	"sync"
)

func simpleWait() {
	var wg sync.WaitGroup
	sayHello := func() {
		defer wg.Done()
		fmt.Println("hello")
	}

	wg.Add(1)
	go sayHello()
	wg.Wait()
}

func main() {
	simpleWait()
}
