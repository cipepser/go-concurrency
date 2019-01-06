package main

import (
	"sync"
	"testing"
)

func BenchmarkContextSwitch(b *testing.B) {
	var wg sync.WaitGroup
	begin := make(chan struct{})
	c := make(chan struct{})

	var token struct{}
	sender := func() {
		defer wg.Done()
		<-begin
		for i := 0; i < b.N; i++ {
			c <- token
		}
	}

	receiver := func() {
		defer wg.Done()
		<-begin
		for i := 0; i < b.N; i++ {
			<-c
		}
	}

	wg.Add(2)
	go sender()
	go receiver()
	b.StartTimer()
	close(begin)
	wg.Wait()

	//❯ go test -bench=. -cpu=1
	//goos: darwin
	//goarch: amd64
	//pkg: github.com/cipepser/go-concurrency/chap3
	//BenchmarkContextSwitch 	10000000	       191 ns/op
	//PASS
	//ok  	github.com/cipepser/go-concurrency/chap3	2.133s
}
