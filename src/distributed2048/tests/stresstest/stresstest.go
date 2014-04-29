package main

import (
	"distributed2048/cmdlineclient"
	"distributed2048/lib2048"
	"distributed2048/util"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var LOGV = util.NewLogger(false, "LIBSTORETEST", os.Stdout)

var (
	numClients = flag.Int("numClients", 1, "number of game clients")
	// numGameServers        = flag.Int("numGameServers", 1, "number of game servers")
	gameServerHostPorts   = flag.String("gsHostPorts", "", "comma separated list of host:port for each game server")
	centralServerHostPort = flag.String("csHostPort", fmt.Sprintf("localhost:%d", util.CENTRALPORT), "host:port of central server")
	numMoves              = flag.Int("numMoves", 10, "number of random moves to generate")
	numSendingClients     = flag.Int("numSendingClients", 1, "number of clients that will be sending moves")
	sendMoveInterval      = flag.Int("sendMoveInterval", 1000, "number of milliseconds between sending moves")
)

type testFunc struct {
	name string
	f    func()
}

func test() bool {
	flag.Parse()
	useCentral := len(*gameServerHostPorts) == 0
	gsHostPortsSlice := strings.Split(*gameServerHostPorts, ",")
	if len(gsHostPortsSlice) == 0 && *centralServerHostPort == "" {
		fmt.Println("ERROR: Invalid configuration. Either specify a central server or a list of game servers.")
		os.Exit(-1)
	}

	// Generate random moves
	moveList := make([]*lib2048.Move, 0)
	for i := 0; i < *numMoves; i++ {
		moveList = append(moveList, util.RandomMove())
	}

	// Start up all command line clients
	cservAddr := "http://" + *centralServerHostPort
	clients := make([]cmdlineclient.Cclient, *numClients)
	gsHostPort := ""
	for i := 0; i < *numClients; i++ {
		if !useCentral {
			gsHostPort = gsHostPortsSlice[i%len(gsHostPortsSlice)]
			cservAddr = ""
		}
		var err error
		clients[i], err = cmdlineclient.NewCClient(cservAddr, gsHostPort, util.DEFAULTINTERVAL)
		if err != nil {
			fmt.Println("FAIL: Command line client could not start:", err)
			return false
		}
	}

	// Test the moves
	for i, m := range moveList {
		fmt.Printf("Sending moves (%d/%d)...\n", i+1, len(moveList))
		for i := 0; i < *numSendingClients; i++ {
			clients[i].InputMove(m.Direction)
		}
		time.Sleep(time.Duration(*sendMoveInterval) * time.Millisecond)
	}

	// Check that all clients have the same set of moves
	time.Sleep(7 * time.Second) // Let all network operations complete
	for i, c := range clients {
		if !c.GetGameState().Equals(clients[0].GetGameState()) {
			fmt.Println("FAIL: Game client", i, "differs in state from game client", 0)
			fmt.Printf("Client %d's state:\n%s\n", i, c.GetGameState().String())

			fmt.Printf("Client %d's state:\n%s\n", 0, clients[0].GetGameState().String())
			return false
		}
	}
	fmt.Println("PASS")
	return true
}

func main() {
	if test() {
		os.Exit(7)
	}
	os.Exit(-1)
}
