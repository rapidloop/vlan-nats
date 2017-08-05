package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	nats "github.com/nats-io/go-nats"
	"github.com/songgao/water"
	"github.com/songgao/water/waterutil"
)

var (
	natsURL = flag.String("n", nats.DefaultURL, "the NATS URL to connect to")
	vlanID  = flag.Uint("v", 0, "a VLAN ID (default 0)")
)

func main() {
	// parse command-line args
	flag.Parse()

	// connect to nats
	nc, err := nats.Connect(*natsURL, nats.Timeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// create a TAP interface
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = fmt.Sprintf("vnats%d", *vlanID)
	ifce, err := water.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// get ethernet address of the interface we just created
	var ownEth net.HardwareAddr
	nifces, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}
	for _, nifce := range nifces {
		if nifce.Name == config.Name {
			ownEth = nifce.HardwareAddr
			break
		}
	}
	if len(ownEth) == 0 {
		log.Fatal("failed to get own ethernet address")
	}

	// sub to our frame address
	subTopic := fmt.Sprintf("vlan.%d.%x", *vlanID, ownEth)
	sub, err := nc.Subscribe(subTopic, func(m *nats.Msg) {
		if _, err := ifce.Write(m.Data); err != nil {
			log.Print(err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	// sub to broadcasts
	broadcastTopic := fmt.Sprintf("vlan.%d", *vlanID)
	bsub, err := nc.Subscribe(broadcastTopic, func(m *nats.Msg) {
		if _, err := ifce.Write(m.Data); err != nil {
			log.Print(err)
		}
	})
	if err != nil {
		log.Fatal(err)
	}
	defer bsub.Unsubscribe()

	// read each frame and publish it to appropriate topic
	var frame [1500]byte
	for {
		// read frame from interface
		n, err := ifce.Read(frame[:])
		if err != nil {
			log.Fatal(err)
		}
		frame2 := frame[:n]

		// the topic to publish to
		dst := waterutil.MACDestination(frame2)
		var pubTopic string
		if waterutil.IsBroadcast(dst) {
			pubTopic = broadcastTopic
		} else {
			pubTopic = fmt.Sprintf("vlan.%d.%x", *vlanID, dst)
		}

		// publish
		if err := nc.Publish(pubTopic, frame2); err != nil {
			log.Fatal(err)
		}
	}
}
