package main

import (
	"distributed2048/centralserver"
	"distributed2048/gameserver"
	"distributed2048/rpc/paxosrpc"
	"distributed2048/util"
	"os"
	"time"
)

var LOGV = util.NewLogger(true, "LIBSTORETEST", os.Stdout)

func main() {
	centralServer, err := centralserver.NewCentralServer(15340, 3)
	if err != nil {
		LOGV.Println("Could not start central server.")
		LOGV.Println(err)
		os.Exit(1)
	}
	time.Sleep(1 * time.Second)

	gameServers := make([]gameserver.GameServer, 0)
	for i := 0; i < 3; i++ {
		gs, err := gameserver.NewGameServer("localhost:15340", "localhost", 15400+i)
		if err != nil {
			LOGV.Printf("Could not start game server %d\n", i)
			LOGV.Println(err)
			os.Exit(1)
		}
		gameServers = append(gameServers, gs)
	}

	moves := []paxosrpc.Move{
		paxosrpc.NewMove(paxosrpc.Up),
		paxosrpc.NewMove(paxosrpc.Up),
		paxosrpc.NewMove(paxosrpc.Down),
		paxosrpc.NewMove(paxosrpc.Left),
	}
	gameServers[0].TestAddVote(moves)

	os.Exit(0)
}
