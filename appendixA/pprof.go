package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	log.SetFlags(log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	go func() {
		goroutine := pprof.Lookup("goroutine")
		for range time.Tick(1 * time.Second) {
			log.Printf("goroutine count: %d\n", goroutine.Count())
		}
	}()

	var blockForever chan struct{}
	for i := 0; i < 10; i++ {
		go func() {
			<-blockForever
		}()
		time.Sleep(500 * time.Millisecond)
	}
	//13:46:41 goroutine count: 5
	//13:46:42 goroutine count: 6
	//13:46:43 goroutine count: 8
	//13:46:44 goroutine count: 10
	//13:46:45 goroutine count: 12
}

