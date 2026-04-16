package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	println("service started")

	seeds := []string{"localhost:9092"}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.DialTimeout(5*time.Second),
		kgo.ConsumerGroup("my-group-identifier"),
		kgo.InstanceID("unique-instance-id"), // статический ID
		kgo.SessionTimeout(10*time.Second),
		kgo.ConsumeTopics("foo"),
	)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	defer cl.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Topic: "foo", Value: []byte("bar")}
	cl.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			fmt.Printf("record had a produce error: %v\n", err)
		}
	})
	wg.Wait()

	if err := cl.ProduceSync(ctx, record).FirstErr(); err != nil {
		fmt.Printf("record had a produce error while synchronously producing: %v\n", err)
	}

	for {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			fmt.Println("errors", errs)
			return
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			fmt.Println(string(record.Value), "from an iterator!")
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				fmt.Println(string(record.Value), "from range inside a callback!")
			}

			p.EachRecord(func(record *kgo.Record) {
				fmt.Println(string(record.Value), "from a second callback!")
			})
		})
	}
}
