package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	println("consumer started")

	seeds := []string{"localhost:9092"}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.DialTimeout(5*time.Second),
		kgo.ConsumerGroup("messages-consumers"),
		kgo.InstanceID("consumer"), // static ID
		kgo.SessionTimeout(10*time.Second),
		kgo.ConsumeTopics("messages"),
	)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	defer cl.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			fmt.Println("errors", errs)

			break
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			fmt.Println("received:", string(record.Value))
		}
	}
}
