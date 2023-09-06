package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	ad "github.com/padiazg/notifier/drivers/amqp"
	wd "github.com/padiazg/notifier/drivers/webhook"
	e "github.com/padiazg/notifier/engine"
	n "github.com/padiazg/notifier/notification"
)

const (
	SomeEvent    n.EventType = "SomeEvent"
	AnotherEvent n.EventType = "AnotherEvent"
)

func main() {
	var (
		// Initialize the notification engine
		notificationEngine = &e.NotificationEngine{
			MQ: ad.NewAMQP10Notifier(ad.Config{
				QueueName: "notifier",
				Address:   "amqp://localhost",
			}),
			Webhook: wd.NewWebhookNotifier(wd.Config{
				Endpoint: "https://localhost:8443/webhook",
				Insecure: true,
			}),
			OnError: func(err error) {
				fmt.Printf("Error sending notification: %v\n", err)
			},
		}

		wg   sync.WaitGroup
		done = make(chan bool)
		c    = make(chan os.Signal, 2)
		err  error
	)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Initialize the message queue notifier

	if err = notificationEngine.MQ.Connect(); err != nil {
		panic(err)
	}
	defer notificationEngine.MQ.Close()

	wg.Add(2)
	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #1")
		notificationEngine.Dispatch(&n.Notification{
			Event: SomeEvent,
			Data:  "simple text data",
		})
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #2")
		notificationEngine.Dispatch(&n.Notification{
			Event: AnotherEvent,
			Data: struct {
				ID   uint
				Name string
			}{ID: 1, Name: "complex data"},
		})
	}()

	wg.Wait()
	close(done)

	for {
		select {
		case <-done:
			fmt.Println("Closing")
			return
		case <-c:
			fmt.Println("Ctrl+C pressed")
			close(done)
		}
	}
}
