package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"sync"
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
	Message string
	self    peer.ID
	Sender  peer.ID
}

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
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

	// fix this for future
	var r io.Reader
	r = rand.New(rand.NewSource(int64(123412451241)))
	prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator with the current time

	listenAddr := fmt.Sprintf("/ip4/127.0.0.1/udp/%d/quic-v1", port)

	h, err := libp2p.New(libp2p.Transport(
		libp2pquic.NewTransport),
		libp2p.ListenAddrStrings(listenAddr),
		libp2p.Identity(prvKey))

	if err != nil {
		panic(err)
	}

	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		panic(err)
	}

	peerChan := initMDNS(h, "asd123")

	for {
		peer := <-peerChan // will block until we discover a peer

		if err := h.Connect(ctx, peer); err != nil {
			fmt.Println("Connection failed:", err)
			continue
		}

		sub, topic, err := JoinGossip(ps, "huone 105")
		if err != nil {
			fmt.Println("Error joining gossip:", err)
			return
		}

		go readLoop(ctx, sub, h.ID(), instanceID)
		go NewMessageSender(ctx, topic, h.ID())

		fmt.Println("end of loop")

	}
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

func readLoop(ctx context.Context, sub *pubsub.Subscription, myId peer.ID, instanceID int) {
	fmt.Println("Luetaan....")
	for {
		msg, err := sub.Next(ctx)
		fmt.Println(msg)
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
	n.PeerChan <- pi
}

func initMDNS(peerhost host.Host, rendezvous string) chan peer.AddrInfo {
	n := &discoveryNotifee{}
	n.PeerChan = make(chan peer.AddrInfo)
	ser := mdns.NewMdnsService(peerhost, rendezvous+"asd", n)
	if err := ser.Start(); err != nil {
		panic(err)
	}
	return n.PeerChan
}
