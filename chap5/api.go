package main

import (
	"context"
	"log"
	"os"
	"sync"

	"golang.org/x/time/rate"
)

type APIConnection struct {
	rateLimiter *rate.Limiter
}

func Open() *APIConnection {
	return &APIConnection{
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1),
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	// do something
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	// do something
	return nil
}

func main() {
	defer log.Printf("Done.")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
			log.Printf("ReadFile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
	// No limit
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 ResolveAddress
	//13:49:35 ReadFile
	//13:49:35 Done.

	// set limit to process 1 request / sec
	//12:10:03 ReadFile
	//12:10:04 ReadFile
	//12:10:05 ReadFile
	//12:10:06 ResolveAddress
	//12:10:07 ResolveAddress
	//12:10:08 ResolveAddress
	//12:10:09 ReadFile
	//12:10:10 ResolveAddress
	//12:10:11 ReadFile
	//12:10:12 ReadFile
	//12:10:13 ResolveAddress
	//12:10:14 ResolveAddress
	//12:10:15 ReadFile
	//12:10:16 ResolveAddress
	//12:10:17 ResolveAddress
	//12:10:18 ReadFile
	//12:10:19 ResolveAddress
	//12:10:20 ReadFile
	//12:10:21 ReadFile
	//12:10:22 ResolveAddress
	//12:10:22 Done.
}
