package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	fmt.Println("producer started")

	seeds := []string{"localhost:9092"}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.DialTimeout(5*time.Second),
		kgo.InstanceID("producer"), // static ID
		kgo.SessionTimeout(10*time.Second),
	)
	if err != nil {
		fmt.Println("error", err)
		return
	}
	defer cl.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if err := ctx.Err(); err != nil {
			fmt.Println("errors", err)

			break
		}

		fmt.Print("Введите строку: ")
		scanner.Scan()
		input := scanner.Text()

		if input == "exit" {
			break
		}

		record := &kgo.Record{Topic: "messages", Value: []byte(input)}
		if err := cl.ProduceSync(ctx, record).FirstErr(); err != nil {
			fmt.Printf("record had a produce error while synchronously producing: %v\n", err)
		}

		fmt.Println("Сообщение отправлено:", input)
	}
}
