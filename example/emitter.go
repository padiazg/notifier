package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ac "github.com/padiazg/notifier/connector/amqp"
	wc "github.com/padiazg/notifier/connector/webhook"
	e "github.com/padiazg/notifier/engine"
	n "github.com/padiazg/notifier/notification"
)

const (
	SomeEvent    n.EventType = "SomeEvent"
	AnotherEvent n.EventType = "AnotherEvent"
)

func main() {
	var (
		engine = (&e.Engine{}).New(&e.Config{
			OnError: func(err error) {
				log.Printf("Error: %s", err.Error())
			},
		})

		wg   sync.WaitGroup
		done = make(chan bool)
		c    = make(chan os.Signal, 2)
	)

	// add a webhook notifier
	webhookId := engine.RegisterNotifier(wc.New(&wc.Config{
		Name:     "Webhook",
		Endpoint: "https://localhost:4443/webhook",
		Insecure: true,
		Headers: map[string]string{
			"Authorization": "Bearer xyz123",
			"X-Portal-Id":   "1234567890",
		},
	}))

	// add an AMQP notifier
	amqpId := engine.RegisterNotifier(ac.New(&ac.Config{
		Name:      "AMQP",
		QueueName: "notifier",
		Address:   "amqp://localhost",
		Protocol:  ac.ProtocolAMQP10,
	}))

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Starting engine...")
	engine.Start()

	// first notification
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #1")
		engine.Dispatch(&n.Notification{
			Event: SomeEvent,
			Data:  "simple text data",
		})
	}()

	// second notification
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #2")
		engine.Dispatch(&n.Notification{
			Event: AnotherEvent,
			Data: struct {
				ID   uint
				Name string
			}{ID: 1, Name: "complex data"},
		})
	}()

	// only to webhook
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #3")
		engine.Dispatch(&n.Notification{
			Event:    AnotherEvent,
			Data:     "only to webhook",
			Channels: []string{webhookId},
		})
	}()

	// only to mq
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #4")
		engine.Dispatch(&n.Notification{
			Event:    AnotherEvent,
			Data:     "only to mq",
			Channels: []string{amqpId},
		})
	}()

	wg.Wait()

	//let's stop the engine
	engine.Stop()
	// give some time to the engine to stop
	time.Sleep(1 * time.Second)

	// let's close the done channel and exit the program
	close(done)

	for {
		select {
		case <-done:
			fmt.Println("Exiting...")
			// engine.Stop()
			return
		case <-c:
			fmt.Println("Ctrl+C pressed")
			close(done)
		}
	}
}
