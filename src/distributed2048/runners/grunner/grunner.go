package main

import (
	"distributed2048/gameserver"
	"distributed2048/libpaxos"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	defaultHostname        = "localhost"
	defaultGameServerPort  = 15510
	defaultCentralHostPort = "localhost:15340"
)

var (
	port             = flag.Int("port", defaultGameServerPort, "port number to listen on")
	centralHostPort  = flag.String("central", defaultCentralHostPort, "host:port of central server")
	hostname         = flag.String("hostname", defaultHostname, "hostname of THIS game server")
	isFaulty         = flag.Bool("faulty", false, "whether this game server sometimes lags")
	faultyPercent    = flag.Int("faultyPercent", 25, "how frequently lag occurs")
	lagDuration      = flag.Int("lagDuration", 15, "how long the lag lasts, in seconds")
	lagPrepare       = flag.Bool("lagPrepare", true, "whether receiving a prepare should be lagged")
	lagAccept        = flag.Bool("lagAccept", true, "whether receiving an accept request should be lagged")
	lagDecide        = flag.Bool("lagDecide", true, "whether receiving a decide request should be lagged ")
	maxLagSlotNumber = flag.Int("maxLagSlotNumber", 15, "maximum slot number to lag until")
)

func actionString(action libpaxos.PaxosAction) string {
	switch action {
	case libpaxos.Prepare:
		return "Prepare"
	case libpaxos.Accept:
		return "Accept"
	case libpaxos.Decide:
		return "Decide"
	}
	return ""
}

func interrupt(id uint32, action libpaxos.PaxosAction, slotNumber uint32) {
	if int(slotNumber) > *maxLagSlotNumber {
		return
	}
	if (*lagPrepare && action == libpaxos.Prepare) ||
		(*lagAccept && action == libpaxos.Accept) ||
		(*lagDecide && action == libpaxos.Decide) {
		if num := rand.Intn(100); num < *faultyPercent {
			fmt.Printf("Lagging game server %d on %s step for %d seconds\n", int(id), actionString(action), *lagDuration)
			time.Sleep(time.Duration(*lagDuration) * time.Second)
		}
	}
}

func main() {
	flag.Parse()
	gs, err := gameserver.NewGameServer(*centralHostPort, *hostname, *port, "/abc")
	if err != nil {
		fmt.Println("Could not create game server.")
		fmt.Println(err)
		os.Exit(1)
	}
	if *isFaulty {
		gs.GetLibpaxos().SetInterruptFunc(interrupt)
	}

	fmt.Printf("Game Server running on %s:%d\n", *hostname, *port)

	// Run the game server forever
	select {}
}
