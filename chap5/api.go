package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

type APIConnection struct {
	networkLimit,
	diskLimit,
	apiLimit RateLimiter
}

func Open() *APIConnection {
	return &APIConnection{
		networkLimit: MultiLimiter(
			rate.NewLimiter(Per(2, time.Second), 2),
			rate.NewLimiter(Per(10, time.Minute), 10),
		),
		diskLimit: MultiLimiter(
			rate.NewLimiter(rate.Limit(1), 1),
		),
		apiLimit: MultiLimiter(
			rate.NewLimiter(Per(3, time.Second), 3),
		),
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := MultiLimiter(a.apiLimit, a.diskLimit).Wait(ctx); err != nil {
		return err
	}
	// do something
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := MultiLimiter(a.apiLimit, a.networkLimit).Wait(ctx); err != nil {
		return err
	}
	// do something
	return nil
}

type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	sort.Slice(limiters, func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	})

	return &multiLimiter{
		limiters: limiters,
	}
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

	// set multi-limit
	//14:20:23 ReadFile
	//14:20:24 ResolveAddress
	//14:20:24 ReadFile
	//14:20:25 ReadFile
	//14:20:25 ReadFile
	//14:20:26 ReadFile
	//14:20:26 ReadFile
	//14:20:27 ReadFile
	//14:20:27 ResolveAddress
	//14:20:28 ResolveAddress
	//14:20:29 ReadFile
	//14:20:35 ReadFile
	//14:20:41 ReadFile
	//14:20:47 ResolveAddress
	//14:20:53 ResolveAddress
	//14:20:59 ResolveAddress
	//14:21:05 ResolveAddress
	//14:21:11 ResolveAddress
	//14:21:17 ResolveAddress
	//14:21:23 ResolveAddress
	//14:21:23 Done.

	// network, disk and api limit
	//14:33:06 ReadFile
	//14:33:06 ResolveAddress
	//14:33:06 ResolveAddress
	//14:33:07 ResolveAddress
	//14:33:07 ReadFile
	//14:33:07 ResolveAddress
	//14:33:08 ResolveAddress
	//14:33:08 ReadFile
	//14:33:08 ResolveAddress
	//14:33:09 ResolveAddress
	//14:33:09 ReadFile
	//14:33:09 ResolveAddress
	//14:33:10 ResolveAddress
	//14:33:10 ResolveAddress
	//14:33:10 ReadFile
	//14:33:11 ReadFile
	//14:33:12 ReadFile
	//14:33:13 ReadFile
	//14:33:14 ReadFile
	//14:33:15 ReadFile
	//14:33:15 Done.
}
