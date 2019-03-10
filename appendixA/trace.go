package main

import (
	"context"
	"log"
	"os"
	"runtime/trace"
)

func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace output file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	ctx := context.Background()
	ctx, task := trace.NewTask(ctx, "makeCoffee")
	defer task.End()
	trace.Log(ctx, "orderID", "1")

	coffee := make(chan bool)

	go func() {
		trace.WithRegion(ctx, "extractCoffee", func() {
			cnt := 0
			for i := 0; i < 1000; i++ {
				cnt++
			}
		})
		coffee <- true
	}()
	<-coffee
}
