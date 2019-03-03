package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func adhocBinding() {
	data := make([]int, 4)

	loopData := func(handleData chan<- int) {
		defer close(handleData)
		for i := range data {
			handleData <- data[i]
		}
	}

	handleData := make(chan int)
	go loopData(handleData)

	for num := range handleData {
		fmt.Println(num)
	}
	//0
	//0
	//0
	//0
}

func lexicalBinding() {
	chanOwner := func() <-chan int {
		results := make(chan int, 5)
		go func() {
			defer close(results)
			for i := 0; i <= 5; i++ {
				results <- i
			}
		}()
		return results
	}

	consumer := func(results <-chan int) {
		for result := range results {
			fmt.Printf("Received: %d\n", result)
		}
		fmt.Println("Done receiving!")
	}

	results := chanOwner()
	consumer(results)

	//Received: 0
	//Received: 1
	//Received: 2
	//Received: 3
	//Received: 4
	//Received: 5
	//Done receiving!
}

func mutexBinding() {
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()

		var buff bytes.Buffer
		for _, b := range data {
			fmt.Fprintf(&buff, "%c", b)
		}
		fmt.Println(buff.String())
	}

	var wg sync.WaitGroup
	wg.Add(2)
	data := []byte("golang")
	go printData(&wg, data[:3])
	go printData(&wg, data[3:])

	wg.Wait()
	//ang
	//gol
}

func leakGoroutine() {
	doWork := func(strings <-chan string) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for s := range strings {
				fmt.Println(s)
			}
		}()

		return completed
	}

	doWork(nil)
	// do something here
	fmt.Println("Done.")

	//Done.
	// * "doWork exited." was not printed.
}

func cancelGoroutine() {
	doWork := func(
		done <-chan interface{},
		strings <-chan string,
	) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(terminated)
			for {
				select {
				case s := <-strings:
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()

		return terminated
	}

	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Canceling doWork goroutine...")
		close(done)
	}()

	<-terminated
	fmt.Println("Done.")
	//Canceling doWork goroutine...
	//doWork exited.
	//Done.
}

func blockGoroutineWriting() {
	newRandStream := func() <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)
			for {
				randStream <- rand.Int()
			}
		}()

		return randStream
	}

	randStream := newRandStream()
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}
	//❯ go run main.go
	//3 random ints:
	//1: 5577006791947779410
	//2: 8674665223082153551
	//3: 6129484611666145821

	// * "newRandStream closure exited." was not printed.
}

func cancelGoroutineWriting() {
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)
			for {
				select {
				case randStream <- rand.Int():
				case <-done:
					return
				}
			}
		}()

		return randStream
	}

	done := make(chan interface{})
	randStream := newRandStream(done)
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}
	close(done)

	time.Sleep(1 * time.Second)
	//3 random ints:
	//1: 5577006791947779410
	//2: 8674665223082153551
	//3: 6129484611666145821
	//newRandStream closure exited.
}

func orPattern() {
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

	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)

	fmt.Printf("done after %v", time.Since(start))
	//done after 1.005296233s
}

func justPrintError() {
	checkStatus := func(
		done <-chan interface{},
		urls ...string,
	) <-chan *http.Response {
		response := make(chan *http.Response)
		go func() {
			defer close(response)
			for _, url := range urls {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println(err)
					continue
				}
				select {
				case <-done:
					return
				case response <- resp:
				}
			}
		}()
		return response
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{"https://www.google.com", "https://badhost"}
	for response := range checkStatus(done, urls...) {
		fmt.Printf("Response: %v\n", response.Status)
	}
	//Response: 200 OK
	//Get https://badhost: dial tcp: lookup badhost: no such host
}

func returnResult() {
	type Result struct {
		Error    error
		Response *http.Response
	}
	checkStatus := func(
		done <-chan interface{},
		urls ...string,
	) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)

			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{
					Error:    err,
					Response: resp,
				}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	errCount := 0
	urls := []string{"a", "https://www.google.com", "b", "c", "d"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v", result.Error)
			errCount++
			if errCount >= 3 {
				fmt.Println("Too many errors, breaking!")
				break
			}
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
	// without errCount
	//Response: 200 OK
	//error: Get https://badhost: dial tcp: lookup badhost: no such host

	// after adding errCount
	//error: Get a: unsupported protocol scheme ""Response: 200 OK
	//error: Get b: unsupported protocol scheme ""error: Get c: unsupported protocol scheme ""Too many errors, breaking!
}

func pipeline() {
	multiply := func(values []int, multiplier int) []int {
		multipliedValues := make([]int, len(values))
		for i, v := range values {
			multipliedValues[i] = v * multiplier
		}
		return multipliedValues
	}
	add := func(values []int, additive int) []int {
		addedValues := make([]int, len(values))
		for i, v := range values {
			addedValues[i] = v + additive
		}
		return addedValues
	}

	ints := []int{1, 2, 3, 4}
	//for _, v := range add(multiply(ints, 2), 1) {
	//	fmt.Println(v)
	//}
	////3
	////5
	////7
	////9

	for _, v := range multiply(add(multiply(ints, 2), 1), 2) {
		fmt.Println(v)
	}
	//6
	//10
	//14
	//18
}

func generatePipeline() {
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intStream := make(chan int, len(integers))
		go func() {
			defer close(intStream)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intStream <- i:
				}
			}
		}()
		return intStream
	}

	multiply := func(
		done <-chan interface{},
		intStream <-chan int,
		multiplier int,
	) <-chan int {
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}
		}()
		return multipliedStream
	}

	add := func(
		done <-chan interface{},
		intStream <-chan int,
		additive int,
	) <-chan int {
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()
		return addedStream
	}

	done := make(chan interface{})
	defer close(done)

	intStream := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 2)

	for v := range pipeline {
		fmt.Println(v)
	}
	//6
	//10
	//14
	//18
}

func repeats() {
	repeat := func(done <-chan interface{}, values ...interface{}) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}

	//repeatFn := func(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
	//	valueStream := make(chan interface{})
	//	go func() {
	//		defer close(valueStream)
	//		for {
	//			select {
	//			case <-done:
	//				return
	//			case valueStream <- fn():
	//			}
	//		}
	//	}()
	//	return valueStream
	//}

	toString := func(done <-chan interface{}, valueStream <-chan interface{}) <-chan string {
		stringStream := make(chan string)
		go func() {
			defer close(stringStream)
			for v := range valueStream {
				select {
				case <-done:
					return
				case stringStream <- v.(string):
				}
			}
		}()
		return stringStream
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

	done := make(chan interface{})
	defer close(done)

	//for num := range take(done, repeat(done, 1), 10) {
	//	fmt.Printf("%v ", num)
	//}
	////1 1 1 1 1 1 1 1 1 1

	//rand := func() interface{} { return rand.Int() }
	//for num := range take(done, repeatFn(done, rand), 10) {
	//	fmt.Println(num)
	//}
	////5577006791947779410
	////8674665223082153551
	////6129484611666145821
	////4037200794235010051
	////3916589616287113937
	////6334824724549167320
	////605394647632969758
	////1443635317331776148
	////894385949183117216
	////2775422040480279449

	var message string
	for token := range toString(done, take(done, repeat(done, "I", "am."), 5)) {
		message += token
	}

	fmt.Printf("message: %s...", message)
	//message: Iam.Iam.I...
}

func orDone(done, c <-chan interface{}) <-chan interface{} {
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

func _tee() {
	repeat := func(done <-chan interface{}, values ...interface{}) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
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

	tee := func(
		done <-chan interface{},
		in <-chan interface{},
	) (_, _ <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})

		go func() {
			defer close(out1)
			defer close(out2)
			for val := range orDone(done, in) {
				var out1, out2 = out1, out2
				for i := 0; i < 2; i++ {
					select {
					case out1 <- val:
						out1 = nil
					case out2 <- val:
						out2 = nil
					}
				}
			}
		}()
		return out1, out2
	}

	done := make(chan interface{})
	defer close(done)

	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 4))
	for val1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", val1, <-out2)
	}
	//out1: 1, out2: 1
	//out1: 2, out2: 2
	//out1: 1, out2: 1
	//out1: 2, out2: 2
}

func _bridge() {
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

	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}

	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
	//0 1 2 3 4 5 6 7 8 9
}

func noContext() {
	locale := func(done <-chan interface{}) (string, error) {
		select {
		case <-done:
			return "", fmt.Errorf("canceled")
		case <-time.After(3 * time.Second):
		}
		return "EN/US", nil
	}

	genGreeting := func(done <-chan interface{}) (string, error) {
		switch locale, err := locale(done); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "hello", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}

	genFarewell := func(done <-chan interface{}) (string, error) {
		switch locale, err := locale(done); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "goodbye", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}

	printGreeting := func(done <-chan interface{}) error {
		greeting, err := genGreeting(done)
		if err != nil {
			return err
		}
		fmt.Printf("%s world!\n", greeting)
		return nil
	}

	printFarewell := func(done <-chan interface{}) error {
		farewell, err := genFarewell(done)
		if err != nil {
			return err
		}
		fmt.Printf("%s world!\n", farewell)
		return nil
	}

	// main
	var wg sync.WaitGroup
	done := make(chan interface{})
	defer close(done)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Wait()
	//hello world!
	//goodbye world!
}

func introduceContext() {
	locale := func(ctx context.Context) (string, error) {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Minute):
		}
		return "EN/US", nil
	}

	genGreeting := func(ctx context.Context) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		switch locale, err := locale(ctx); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "hello", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}

	genFarewell := func(ctx context.Context) (string, error) {
		switch locale, err := locale(ctx); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "goodbye", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}

	printGreeting := func(ctx context.Context) error {
		greeting, err := genGreeting(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%s world!\n", greeting)
		return nil
	}

	printFarewell := func(ctx context.Context) error {
		farewell, err := genFarewell(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%s world!\n", farewell)
		return nil
	}

	// main
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
	//cannot print greeting: context deadline exceeded
	//cannot print farewell: context canceled
}

func deadline() {
	locale := func(ctx context.Context) (string, error) {
		if dl, ok := ctx.Deadline(); ok {
			if dl.Sub(time.Now().Add(1*time.Minute)) <= 0 {
				return "", context.DeadlineExceeded
			}
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(1 * time.Minute):
		}
		return "EN/US", nil
	}

	genGreeting := func(ctx context.Context) (string, error) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		switch locale, err := locale(ctx); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "hello", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}

	genFarewell := func(ctx context.Context) (string, error) {
		switch locale, err := locale(ctx); {
		case err != nil:
			return "", err
		case locale == "EN/US":
			return "goodbye", nil
		}
		return "", fmt.Errorf("unsupported locale")
	}

	printGreeting := func(ctx context.Context) error {
		greeting, err := genGreeting(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%s world!\n", greeting)
		return nil
	}

	printFarewell := func(ctx context.Context) error {
		farewell, err := genFarewell(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%s world!\n", farewell)
		return nil
	}

	// main
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
	//cannot print greeting: context deadline exceeded
	//cannot print farewell: context canceled

	// This result is returned earlier than introduceContext()'s one.
}

func main() {
	//adhocBinding()
	//lexicalBinding()
	//mutexBinding()

	//leakGoroutine()
	//cancelGoroutine()

	//blockGoroutineWriting()
	//cancelGoroutineWriting()

	//orPattern()

	//justPrintError()
	//returnResult()

	//pipeline()
	//generatePipeline()

	//repeats()

	//_orDone()

	//_tee()

	//_bridge()

	//noContext()
	//introduceContext()
	deadline()
}
