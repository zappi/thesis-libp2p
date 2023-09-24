package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func NewMessageSender(ctx context.Context, topic *pubsub.Topic, id peer.ID) error {
	fmt.Println("sending daataaa")
	for {
		message := generateRandomMessage(id)
		// fmt.Println("Sending message:", message) // Add this line

		msgBytes, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}

		return topic.Publish(ctx, msgBytes)
	}
}

func generateRandomMessage(id peer.ID) SingleMessage {
	// Generate a random message (you can customize this logic)
	message := SingleMessage{
		Message: fmt.Sprintf("Random message at %s", time.Now().Format("2006-01-02 15:04:05")),
		self:    id,
		Sender:  id,
	}
	return message
}
