package main

import (
	"distributed2048/centralserver"
	"flag"
	"fmt"
	"os"
)

const defaultCentralServerPort = 25340

var (
	port           = flag.Int("port", defaultCentralServerPort, "port number to listen on")
	numGameServers = flag.Int("gameservers", 1, "the number of game servers in the cluster")
)

func main() {
	flag.Parse()
	_, err := centralserver.NewCentralServer(*port, *numGameServers)
	if err != nil {
		fmt.Println("Could not create central server.")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Central Server running on port", *port)

	// Run the central server forever
	select {}
}
