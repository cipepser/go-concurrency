package main

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net"
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

func cond() {
	c := sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		c.L.Unlock()
		c.Signal()
	}

	for i := 0; i < 10; i++ {
		c.L.Lock()
		for len(queue) == 2 {
			c.Wait()
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}

	//Adding to queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
	//Removed from queue
	//Adding to queue
}

func broadcast() {
	type Button struct {
		Clicked *sync.Cond
	}
	button := Button{
		Clicked: sync.NewCond(&sync.Mutex{}),
	}

	subscribe := func(c *sync.Cond, fn func()) {
		var goroutineRunning sync.WaitGroup
		goroutineRunning.Add(1)
		go func() {
			goroutineRunning.Done()
			c.L.Lock()
			defer c.L.Unlock()
			c.Wait()
			fn()
		}()
		goroutineRunning.Wait()
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing windows.")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialog box!")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse clicked.")
		clickRegistered.Done()
	})

	button.Clicked.Broadcast()
	// Signal()はruntimeが管理するgoroutineのFIFOを見て、先頭のgoroutineに伝える
	// 今回のようにBroadcast()だと全員に伝える。

	clickRegistered.Wait()

	//Mouse clicked.
	//Maximizing windows.
	//Displaying annoying dialog box!
}

func once() {
	var count int

	increment := func() { count++ }
	decrement := func() { count-- }

	var once sync.Once
	once.Do(increment)
	once.Do(decrement)

	fmt.Printf("Count is %d\n", count) // Count is 1   (not 0)
}

func onceWithAnotherFunc() {
	var count int

	increment := func() {
		count++
	}

	var once sync.Once

	var increments sync.WaitGroup
	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			once.Do(increment)
		}()
	}

	increments.Wait()
	fmt.Printf("Count is %d\n", count) // Count is 1

}

func pool() {
	myPool := &sync.Pool{
		New: func() interface{} {
			fmt.Println("Creating new instance.")
			return struct{}{}
		},
	}

	myPool.Get()
	instance := myPool.Get()
	myPool.Put(instance)
	myPool.Get()

	//Creating new instance.
	//Creating new instance.
}

func poolWithSmallMemory() {
	var numCalcsCreated int
	calcPool := &sync.Pool{
		New: func() interface{} {
			numCalcsCreated++
			mem := make([]byte, 1024)
			return &mem
		},
	}
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())
	calcPool.Put(calcPool.New())

	const numWorkers = 1024 * 1024
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := numWorkers; i > 0; i-- {
		go func() {
			defer wg.Done()

			mem := calcPool.Get().(*[]byte)
			defer calcPool.Put(mem)
		}()
	}

	wg.Wait()
	fmt.Printf("%d calculators were created.", numCalcsCreated) // 4 calculators were created.
}

func connectToservice() interface{} {
	time.Sleep(1 * time.Second)
	return struct{}{}
}

func startNetworkDaemon() *sync.WaitGroup {
	connPool := warmServiceConnCache()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		server, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Fatalf("cannot listen: %v", err)
		}
		defer server.Close()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("connot accept connection: %v", err)
				continue
			}
			svcConn := connPool.Get()
			fmt.Fprintf(conn, "")
			connPool.Put(svcConn)
			conn.Close()
		}
	}()
	return &wg
}

func warmServiceConnCache() *sync.Pool {
	p := &sync.Pool{
		New: connectToservice,
	}

	for i := 0; i < 10; i++ {
		p.Put(p.New())
	}
	return p
}

func deadlockWithChannel() {
	stringStream := make(chan string)
	go func() {
		if 0 != 1 {
			return
		}
		stringStream <- "Hello Channels!"
	}()
	fmt.Println(<-stringStream)

	// doesn't work well in this file, the following result is in scratch file.
	//
	//GOROOT=/usr/local/Cellar/go/1.11.2/libexec #gosetup
	//GOPATH=.go #gosetup
	///usr/local/Cellar/go/1.11.2/libexec/bin/go build -o /private/var/folders/mc/3v_pttq16pdblh7vbqf4mvk80000gn/T/___go_build_scratch_4_go /Library/Preferences/GoLand2018.3/scratches/scratch_4.go #gosetup
	///private/var/folders/mc/3v_pttq16pdblh7vbqf4mvk80000gn/T/___go_build_scratch_4_go #gosetup
	//fatal error: all goroutines are asleep - deadlock!
	//
	//goroutine 1 [chan receive]:
	//main.main()
	//	/Library/Preferences/GoLand2018.3/scratches/scratch_4.go:13 +0x7c
	//
	//Process finished with exit code 2
}

func receiveWithOption() {
	stringStream := make(chan string)
	go func() {
		stringStream <- "Hello Channels!"
	}()
	salutation, ok := <-stringStream
	fmt.Printf("(%v): %v", ok, salutation)

	//(true): Hello Channels!
}

func receiveFromClosedChannel() {
	intStream := make(chan int)
	close(intStream)

	integer, ok := <-intStream
	fmt.Printf("(%v): %v", ok, integer)
	// (false): 0
}

func rangeForChannel() {
	intStream := make(chan int)

	go func() {
		defer close(intStream)
		for i := 0; i < 5; i++ {
			intStream <- i
		}
	}()

	for integer := range intStream {
		fmt.Printf("%v ", integer)
	}

	//0 1 2 3 4
}

func releaseMultiGoroutines() {
	begin := make(chan interface{})
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-begin
			fmt.Printf("%v has begun\n", i)
		}(i)
	}

	fmt.Println("Unblocking goroutines...")
	close(begin) // Without this statement, channel blocks forever, we can see only `Unblocking goroutines...`
	wg.Wait()

	//Unblocking goroutines...
	//0 has begun
	//2 has begun
	//3 has begun
	//1 has begun
}

func bufferedChannel() {
	var stdoutBuff bytes.Buffer
	defer stdoutBuff.WriteTo(os.Stdout)

	intStream := make(chan int, 4)
	go func() {
		defer close(intStream)
		defer fmt.Fprintf(&stdoutBuff, "Producer Done.\n")

		for i := 0; i < 5; i++ {
			fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
			intStream <- i
		}
	}()

	for integer := range intStream {
		fmt.Fprintf(&stdoutBuff, "Received %v.\n", integer)
	}

	//Sending: 0
	//Sending: 1
	//Sending: 2
	//Sending: 3
	//Sending: 4
	//Producer Done.
	//Received 0.
	//Received 1.
	//Received 2.
	//Received 3.
	//Received 4.
}

func nilChannel() {
	var dataStream chan interface{}

	<-dataStream
	// doesn't work well in this file, the following result is in scratch file.
	//
	//fatal error: all goroutines are asleep - deadlock!
	//
	//goroutine 1 [chan receive (nil chan)]:
	//main.main()
	//
	//Process finished with exit code 2

	dataStream <- struct{}{}
	// doesn't work well in this file, the following result is in scratch file.
	//
	//fatal error: all goroutines are asleep - deadlock!
	//
	//goroutine 1 [chan send (nil chan)]:
	//main.main()
	//
	//Process finished with exit code 2

	close(dataStream)
	// doesn't work well in this file, the following result is in scratch file.
	//
	//panic: close of nil channel
	//
	//goroutine 1 [running]:
	//main.main()
	//	/Users/respepic/Library/Preferences/GoLand2018.3/scratches/scratch_5.go:8 +0x2a
	//
	//Process finished with exit code 2
}

func ownedChannel() {
	chanOwner := func() <-chan int {
		resultStream := make(chan int, 5)
		go func() {
			defer close(resultStream)
			for i := 0; i <= 5; i++ {
				resultStream <- i
			}
		}()
		return resultStream
	}

	resultStream := chanOwner()
	for result := range resultStream {
		fmt.Printf("Received: %d\n", result)
	}
	fmt.Println("Done Receiving!")

	//Received: 0
	//Received: 1
	//Received: 2
	//Received: 3
	//Received: 4
	//Received: 5
	//Done Receiving!
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
	//rwmutex()

	//cond()
	//broadcast()

	//once()
	//onceWithAnotherFunc()

	//pool()
	//poolWithSmallMemory()

	//deadlockWithChannel()

	//receiveWithOption()
	//receiveFromClosedChannel()

	//rangeForChannel()

	//releaseMultiGoroutines()

	//bufferedChannel()

	//nilChannel()

	ownedChannel()
}
