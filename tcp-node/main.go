package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

type SingleMessage struct {
	Message string
	self    peer.ID
	Sender  peer.ID
}

type discoveryNotifee struct {
	h host.Host
}

func main() {
	numInstances := 1
	wg := sync.WaitGroup{}

	var port int
	flag.IntVar(&port, "port", 0, "Port number")
	flag.Parse()

	for i := 0; i < numInstances; i++ {
		wg.Add(1)
		go startInstance(&wg, i, port)
	}

	wg.Wait()
	fmt.Println("All instances have completed.")
}

func startInstance(wg *sync.WaitGroup, instanceID int, port int) {
	defer wg.Done()
	fmt.Printf("Instance %d started\n", instanceID)

	ctx := context.Background()

	// prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	// if err != nil {
	// 	panic(err)
	// }

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator with the current time

	listenAddr := fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/", port)

	h, err := libp2p.New(libp2p.ListenAddrStrings(listenAddr))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	// libp2pHost, err := libp2p.New(
	// 	libp2p.ListenAddrs(sourceMultiAddr),
	// 	libp2p.Identity(prvKey),
	// )
	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	if err := setupDiscovery(h); err != nil {
		panic(err)
	}

	sub, topic, err := JoinGossip(ps, "huone 105")
	if err != nil {
		panic(err)
	}

	go readLoop(ctx, sub, h.ID(), instanceID)
	for {
		go NewMessageSender(ctx, topic, h.ID())
		time.Sleep(generateRandomInterval())
	}
	fmt.Printf("Instance %d completed\n", instanceID)
}

func generateRandomInterval() time.Duration {

	minInterval := 30 * time.Second
	maxInterval := 60 * time.Second
	return minInterval + time.Duration(rand.Int63n(int64(maxInterval-minInterval)))
}

func JoinGossip(ps *pubsub.PubSub, roomName string) (*pubsub.Subscription, *pubsub.Topic, error) {
	topic, err := ps.Join(roomName)
	if err != nil {
		panic(err)
	}
	defer topic.Close()
	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

	return sub, topic, err
}

// func (messages *GossipMessageRoom) ListPeers() []peer.ID {
// 	fmt.Println("list called")
// 	return messages.ps.ListPeers("topicName")
// }

func readLoop(ctx context.Context, sub *pubsub.Subscription, myId peer.ID, instanceID int) {
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("Reading loop error:", err)
			panic(err)
		}

		sm := new(SingleMessage)
		err = json.Unmarshal(msg.Data, sm)

		if sm.Sender == myId {
			continue
		}

		if err != nil {
			panic(err)

		}
		fmt.Println("Instance: ", instanceID, "reading message", sm.Message)
	}
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
}

func setupDiscovery(h host.Host) error {
	// setup mDNS discovery to find local peers
	s := mdns.NewMdnsService(h, "pubsub-example", &discoveryNotifee{h: h})
	return s.Start()
}
