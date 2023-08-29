package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/padiazg/notifier"
)

const (
	UserCreated notifier.EventType = "UserCreated"
	UserUpdated notifier.EventType = "UserUpdated"
	UserDeleted notifier.EventType = "UserDeleted"
)

func main() {
	var (
		// Initialize the notification engine
		notificationEngine = &notifier.NotificationEngine{
			Webhook: &notifier.WebhookNotifier{
				Endpoint: "https://localhost:8443/webhook",
				Insecure: true,
			},
			MQ: &notifier.AMQP10Notifier{},
			OnError: func(err error) {
				fmt.Printf("Error sending notification: %v\n", err)
			},
		}

		wg   sync.WaitGroup
		done = make(chan bool)
		c    = make(chan os.Signal, 2)
	)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	wg.Add(2)
	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #1")
		notificationEngine.Dispatch(&notifier.Notification{
			Event: UserCreated,
			Data:  "sample data #1",
		})
	}()

	go func() {
		defer wg.Done()
		fmt.Println("Sending notification #2")
		notificationEngine.Dispatch(&notifier.Notification{
			Event: UserUpdated,
			Data: struct {
				ID   uint
				Name string
			}{ID: 1, Name: "sample data #2"},
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
