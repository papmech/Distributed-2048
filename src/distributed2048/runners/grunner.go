package main

import (
	"distributed2048/gameserver"
	"flag"
	"fmt"
	"os"
)

const (
	defaultHostname        = "localhost"
	defaultGameServerPort  = 15510
	defaultCentralHostPort = "localhost:15340"
)

var (
	port            = flag.Int("port", defaultGameServerPort, "port number to listen on")
	centralHostPort = flag.String("central", defaultCentralHostPort, "host:port of central server")
	hostname        = flag.String("hostname", defaultHostname, "hostname of THIS game server")
)

func main() {
	flag.Parse()
	_, err := gameserver.NewGameServer(*centralHostPort, *hostname, *port, "/")
	if err != nil {
		fmt.Println("Could not create game server.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Game Server running on %s:%d\n", *hostname, *port)

	// Run the game server forever
	select {}
}
