package main

import (
	"io/ioutil"
	"net"
	"testing"
)

//func BenchmarkContextSwitch(b *testing.B) {
//	var wg sync.WaitGroup
//	begin := make(chan struct{})
//	c := make(chan struct{})
//
//	var token struct{}
//	sender := func() {
//		defer wg.Done()
//		<-begin
//		for i := 0; i < b.N; i++ {
//			c <- token
//		}
//	}
//
//	receiver := func() {
//		defer wg.Done()
//		<-begin
//		for i := 0; i < b.N; i++ {
//			<-c
//		}
//	}
//
//	wg.Add(2)
//	go sender()
//	go receiver()
//	b.StartTimer()
//	close(begin)
//	wg.Wait()
//
//	//❯ go test -bench=. -cpu=1
//	//goos: darwin
//	//goarch: amd64
//	//pkg: github.com/cipepser/go-concurrency/chap3
//	//BenchmarkContextSwitch 	10000000	       191 ns/op
//	//PASS
//	//ok  	github.com/cipepser/go-concurrency/chap3	2.133s
//}

func init() {
	daemonStarted := startNetworkDaemon()
	daemonStarted.Wait()
}

func BenchmarkNetworkRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			b.Fatalf("cannot dial host: %v", err)
		}
		if _, err := ioutil.ReadAll(conn); err != nil {
			b.Fatalf("cannot read: %v", err)
		}
		conn.Close()
	}

	// *** No Pool ***
	//❯ go test -benchtime=10s -bench=.
	//goos: darwin
	//goarch: amd64
	//pkg: github.com/cipepser/go-concurrency/chap3
	//BenchmarkNetworkRequest-4   	      10	1004729582 ns/op
	//PASS
	//ok  	github.com/cipepser/go-concurrency/chap3	11.063s

	// *** With Pool ***
	//❯ go test -benchtime=10s -bench=.
	//goos: darwin
	//goarch: amd64
	//pkg: github.com/cipepser/go-concurrency/chap3
	//BenchmarkNetworkRequest-4   	    1000	  10389753 ns/op
	//PASS
	//ok  	github.com/cipepser/go-concurrency/chap3	25.766s
}
