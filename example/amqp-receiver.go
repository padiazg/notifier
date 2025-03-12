package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Azure/go-amqp"
	n "github.com/padiazg/notifier/notification"
)

func main() {
	var (
		ctx       = context.TODO()
		conn, err = amqp.Dial(context.TODO(), "amqp:localhost", nil)
	)

	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer conn.Close()

	session, err := conn.NewSession(context.TODO(), nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)

	}

	{
		// create a receiver
		receiver, err := session.NewReceiver(ctx, "notifier", &amqp.ReceiverOptions{
			SourceDurability: amqp.DurabilityUnsettledState,
		})
		if err != nil {
			log.Fatal("Creating receiver link:", err)
		}

		defer func() {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			receiver.Close(ctx)
			cancel()
		}()

		log.Printf("Connected to queue\n")

		for {
			// receive next message
			msg, err := receiver.Receive(ctx, nil)
			if err != nil {
				log.Printf("Reading message from AMQP: %v\n", err)
			}

			// accept message
			if err = receiver.AcceptMessage(context.TODO(), msg); err != nil {
				log.Printf("Failure accepting message: %v", err)
			}

			// Parse the incoming JSON payload
			var notification n.Notification
			err = json.Unmarshal(msg.GetData(), &notification)
			if err != nil {
				log.Printf("Failed to decode JSON payload: %v", err)
				return
			}

			formated, err := json.MarshalIndent(notification, "", "  ")
			if err != nil {
				log.Printf("Failed to format payload: %v", err)
				return
			}

			log.Printf("Message received: %s\n", formated)
		}
	}

}
