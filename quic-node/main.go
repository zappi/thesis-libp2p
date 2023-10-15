package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
)

type SingleMessage struct {
	Message  string
	self     peer.ID
	Sender   peer.ID
	NodeName string
}

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port_number>")
		os.Exit(1)
	}
	portStr := os.Args[1]
	nodeName := os.Args[2]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println("Invalid port number:", err)
		os.Exit(1)
	}

	fmt.Printf("Selected port: %d\n", port)

	startInstance(port, nodeName)
}

func startInstance(port int, nodeName string) {

	ctx := context.Background()

	// fix this for future
	var r io.Reader
	r = rand.New(rand.NewSource(int64(123412451241)))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())

	listenAddr := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", port)

	fmt.Println(listenAddr)

	h, err := libp2p.New(libp2p.Transport(
		libp2pquic.NewTransport),
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.Identity(prvKey))

	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithMaxMessageSize(4096))
	if err != nil {
		panic(err)
	}

	peerChan := initMDNS(h, "mdns-rendezvous")

	for {
		peer := <-peerChan

		if err := h.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}

		sub, topic, err := JoinGossip(ps, "topic1")
		if err != nil {
			fmt.Println("Error joining gossip:", err)
			return
		}

		go readLoop(ctx, sub, h.ID())
		go messageSender(ctx, topic, h.ID(), nodeName)
		go sendTimestampPeriodically(ctx, topic, h.ID(), nodeName)
	}

}

func sendTimestampPeriodically(ctx context.Context, topic *pubsub.Topic, id peer.ID, nodeName string) {
	for {
		randomInterval := time.Duration(rand.Intn(16)+5) * time.Second

		time.Sleep(randomInterval)

		timestampMessage := SingleMessage{
			Message:  time.Now().Format("2006-01-02 15:04"),
			self:     id,
			Sender:   id,
			NodeName: nodeName,
		}
		sendGeneratedMessage(timestampMessage, ctx, topic, id)
	}
}

func JoinGossip(ps *pubsub.PubSub, topicName string) (*pubsub.Subscription, *pubsub.Topic, error) {
	topic, err := ps.Join(topicName)
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

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("Discovered new peer %s\n", pi.ID.Pretty())
	n.PeerChan <- pi
}

func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	notifee := &discoveryNotifee{}
	notifee.PeerChan = make(chan peer.AddrInfo)
	mdnsService := mdns.NewMdnsService(peerhost, rendezvous, notifee)
	if err := mdnsService.Start(); err != nil {
		panic(err)
	}
	return notifee.PeerChan
}
