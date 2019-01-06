package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"text/tabwriter"
	"time"
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

func sizeOfGoroutine() {
	memConsumed := func() uint64 {
		runtime.GC()
		var s runtime.MemStats
		runtime.ReadMemStats(&s)
		return s.Sys
	}

	var c <-chan interface{}
	var wg sync.WaitGroup
	noop := func() { wg.Done(); <-c }

	const numGoroutines = 1e4
	wg.Add(numGoroutines)
	before := memConsumed()
	for i := numGoroutines; i > 0; i-- {
		go noop()
	}
	wg.Wait()
	after := memConsumed()
	fmt.Printf("%.3fkb", float64(after-before)/numGoroutines/1000)
	// 0.157kb
	// 0.000kbとかにもなった。。。
}

func waitgroup() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2nd goroutine sleeping...")
		time.Sleep(2)
	}()

	wg.Wait()
	fmt.Println("All goroutines complete.")

	//2nd goroutine sleeping...
	//1st goroutine sleeping...
	//All goroutines complete.
}

func waitgroupWithLoop() {
	hello := func(wg *sync.WaitGroup, id int) {
		defer wg.Done()
		fmt.Printf("Hello from %v!\n", id)
	}

	const numGreeters = 5
	var wg sync.WaitGroup
	wg.Add(numGreeters)
	for i := 0; i < numGreeters; i++ {
		go hello(&wg, i+1)
	}
	wg.Wait()

	//Hello from 5!
	//Hello from 4!
	//Hello from 3!
	//Hello from 1!
	//Hello from 2!
}

func mutex() {
	var count int
	var lock sync.Mutex

	increment := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
		fmt.Printf("Incrementing: %d\n", count)
	}

	decrement := func() {
		lock.Lock()
		defer lock.Unlock()
		count--
		fmt.Printf("Decrementing: %d\n", count)
	}

	var arithmetic sync.WaitGroup
	for i := 0; i < 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			increment()
		}()
	}
	for i := 0; i < 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			decrement()
		}()
	}

	arithmetic.Wait()

	fmt.Println("Arithmetic complete.")
	//Incrementing: 1
	//Incrementing: 2
	//Incrementing: 3
	//Decrementing: 2
	//Incrementing: 3
	//Decrementing: 2
	//Decrementing: 1
	//Decrementing: 0
	//Decrementing: -1
	//Incrementing: 0
	//Arithmetic complete.
}

func rwmutex() {
	producer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			l.Lock()
			l.Unlock()
			time.Sleep(1)
		}
	}

	observer := func(wg *sync.WaitGroup, l sync.Locker) {
		defer wg.Done()
		l.Lock()
		defer l.Unlock()
	}

	test := func(count int, mutex, rmMutex sync.Locker) time.Duration {
		var wg sync.WaitGroup
		wg.Add(count + 1)
		beginTestTime := time.Now()
		go producer(&wg, mutex)
		for i := count; i > 0; i-- {
			go observer(&wg, rmMutex)
		}

		wg.Wait()
		return time.Since(beginTestTime)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	defer tw.Flush()

	var m sync.RWMutex
	fmt.Fprintf(tw, "Readers\tRWMutex\tMutex\n")
	for i := 0; i < 20; i++ {
		count := int(math.Pow(2, float64(i)))
		fmt.Fprintf(
			tw,
			"%d\t%v\t%v\n",
			count,
			test(count, &m, m.RLocker()),
			test(count, &m, &m),
		)
	}

	//Readers  RWMutex       Mutex
	//1        31.826µs      6.123µs
	//2        35.848µs      69.869µs
	//4        15.426µs      20.941µs
	//8        24.27µs       39.03µs
	//16       57.93µs       67.618µs
	//32       76.532µs      78.665µs
	//64       107.06µs      122.535µs
	//128      103.991µs     248.311µs
	//256      249.383µs     117.04µs
	//512      207.498µs     215.452µs
	//1024     333.15µs      510.055µs
	//2048     806.2µs       767.582µs
	//4096     1.188948ms    1.206972ms
	//8192     2.786535ms    3.487278ms
	//16384    7.557565ms    17.382744ms
	//32768    28.781047ms   48.623299ms
	//65536    81.844755ms   42.963367ms
	//131072   36.88194ms    38.454013ms
	//262144   68.751632ms   69.51097ms
	//524288   147.760192ms  150.12134ms

	// 実行するたびに、RWMutexとMutexの速さが変わるので他のアプリ落としたり、
	// 環境を整えないと正しい結果にならなそう。
	// それくらいで状況が変わってしまうならRWMutexを使う理由がないような気もする。
}

func main() {
	//simpleWait()

	//scope()

	//badRangeExpression()
	//goodRangeExpression()

	//sizeOfGoroutine()

	//waitgroup()
	//waitgroupWithLoop()

	//mutex()
	rwmutex()
}
