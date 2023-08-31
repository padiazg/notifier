package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/go-amqp"
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
				log.Fatal("Reading message from AMQP:", err)
			}

			// accept message
			if err = receiver.AcceptMessage(context.TODO(), msg); err != nil {
				log.Fatalf("Failure accepting message: %v", err)
			}

			fmt.Printf("Message received: %s\n", msg.GetData())
		}
	}

}
