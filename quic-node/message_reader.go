package main

import (
	"context"
	"encoding/json"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (sm *SingleMessage) String() string {
	jsonBytes, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}
	return string(jsonBytes)
}

func readLoop(ctx context.Context, sub *pubsub.Subscription, myID peer.ID) {
	fmt.Println("Read loop started")

	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Error reading message:", err)
			continue
		}

		sm := new(SingleMessage)

		err = json.Unmarshal(msg.Data, sm)

		if sm.Sender == myID {
			continue
		}

		if err != nil {
			fmt.Println("Error decoding message:", err)
			continue
		}

		fmt.Printf("Received message: %+v\n", sm)
	}
}
