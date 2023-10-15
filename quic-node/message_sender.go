package main

import (
	"context"
	"encoding/json"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func messageSender(ctx context.Context, topic *pubsub.Topic, id peer.ID, nodeName string) error {
	for {
		var userInput string
		fmt.Scanln(&userInput)

		message := SingleMessage{
			userInput,
			id,
			id,
			nodeName,
		}

		msgBytes, err := json.Marshal(message)
		if err != nil {
			fmt.Println("Error marshaling message:", err)
			continue
		}

		err = topic.Publish(ctx, msgBytes)
		if err != nil {
			fmt.Println("Error publishing message:", err)
			continue
		}
	}
}

func sendGeneratedMessage(generatedMessage SingleMessage, ctx context.Context, topic *pubsub.Topic, id peer.ID) {
	msgBytes, err := json.Marshal(generatedMessage)
	if err != nil {
		fmt.Println("Error encoding timestamp message:", err)
	}

	err = topic.Publish(ctx, msgBytes)
	if err != nil {
		fmt.Println("Error publishing timestamp message:", err)
	}

}
