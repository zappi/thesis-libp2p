package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func messageSender(ctx context.Context, topic *pubsub.Topic, id peer.ID, nodeName string) error {
	reader := bufio.NewReader(os.Stdin)

	for {
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim spaces and newlines from the input
		userInput = strings.TrimSpace(userInput)

		message := SingleMessage{
			Message:  userInput,
			self:     id,
			Sender:   id,
			NodeName: nodeName,
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
