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

func scope() {
	var wg sync.WaitGroup
	salutation := "hello"

	wg.Add(1)
	go func() {
		defer wg.Done()
		salutation = "welcome"
	}()

	wg.Wait()
	fmt.Println(salutation) // welcome
}

func badRangeExpression() {
	var wg sync.WaitGroup

	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(salutation)
		}()
	}

	wg.Wait()

	//good day
	//good day
	//good day
}

func goodRangeExpression() {
	var wg sync.WaitGroup

	for _, salutation := range []string{"hello", "greetings", "good day"} {
		wg.Add(1)
		go func(salutation string) {
			defer wg.Done()
			fmt.Println(salutation)
		}(salutation)
	}

	wg.Wait()

	//good day
	//greetings
	//hello
}

func main() {
	//simpleWait()

	//scope()

	//badRangeExpression()
	goodRangeExpression()
}
