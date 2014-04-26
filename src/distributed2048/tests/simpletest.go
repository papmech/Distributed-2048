package main

// ======================================================================== //
// Dis be berry simple test
// Dere be 1 central server + 1 game server + 1 client
// Der client be making 3 simple moves
// ======================================================================== //
import (
	"distributed2048/centralserver"
	"distributed2048/gameserver"
	"distributed2048/rpc/paxosrpc"
	"distributed2048/cmdlineclient"
	"distributed2048/util"
	"os"
	"time"
)


var LOGV = util.NewLogger(true, "LIBSTORETEST", os.Stdout)


func main() {
	// Step 1: Boot Central Server
	_, err := centralserver.NewCentralServer(util.CENTRALPORT, 1)
	if err != nil {
		LOGV.Println(err)
		LOGV.Println(util.CSFAIL)
		os.Exit(1)
	}
	time.Sleep(3 * time.Second)

	// Step 2: Boot Game Server
	_, err := gameserver.NewGameServer(util.CENTRALHOSTPOST, util.LOCALHOST,
		util.GAMESERVERPORT, util.DEFAULTPATTERN)
	if err != nil {
		LOGV.Println(err)
		LOGV.Println(util.GSFAIL)
		os.Exit(1)
	}
	time.Sleep(3 * time.Second)

	// Step 3: Boot Testing Client
	cservAddr := "http://" + util.CENTRALHOSTPOST
	cli, err := cmdlineclient.NewCClient(cservAddr, util.DEFAULTINTERVAL)
	if err != nil {
		LOGV.Println(err)
		LOGV.Println(util.CFAIL)
		os.Exit(1)
	}
	time.Sleep(3 * time.Second)

	// Initialize moves + obtain correct answer
	movelist := []int{ paxosrpc.Left, paxosrpc.Right, paxosrpc.Left }


	// Step 4: Obtain Right Answer

}
