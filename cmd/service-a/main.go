package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	println("service-a started")

	seeds := []string{"localhost:9092"}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.DialTimeout(5*time.Second),
		kgo.ConsumerGroup("my-group-identifier"),
		kgo.ConsumeTopics("foo"),
		// kgo.AutoCommitMarks(),
	)
	if err != nil {
		panic(err)
	}
	defer cl.Close()

	ctx := context.Background()

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

	var i int
	for {
		fetches := cl.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			panic(fmt.Sprint(errs))
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			fmt.Println(string(record.Value), "from an iterator!")
			i++
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				fmt.Println(string(record.Value), "from range inside a callback!")
				i++
			}

			p.EachRecord(func(record *kgo.Record) {
				fmt.Println(string(record.Value), "from a second callback!")
				i++
			})
		})

		cl.CommitUncommittedOffsets(ctx)

		if i >= 6 {
			break
		}
	}
}
