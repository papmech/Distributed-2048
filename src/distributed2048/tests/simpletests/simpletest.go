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
	"distributed2048/tests"
	"os"
	"time"
)


var LOGV = util.NewLogger(true, "LIBSTORETEST", os.Stdout)

func processError(err error, msg string, pause int) {
	if err != nil {
		LOGV.Println(err)
		LOGV.Println(msg)
		os.Exit(1)
	}
	if pause > 0 {
		time.Sleep(time.Duration(pause) * time.Second)
	}
}

func main() {
	// Step 1: Boot Central Server
	_, err := centralserver.NewCentralServer(util.CENTRALPORT, 1)
	processError(err, util.CSFAIL, 3)

	// Step 2: Boot Game Server
	_, err = gameserver.NewGameServer(util.CENTRALHOSTPOST, util.LOCALHOST, util.GAMESERVERPORT, util.DEFAULTPATTERN)
	processError(err, util.GSFAIL, 3)

	// Step 3: Boot Testing Client
	cservAddr := "http://" + util.CENTRALHOSTPOST
	cli, err := cmdlineclient.NewCClient(cservAddr, util.DEFAULTINTERVAL)
	processError(err, util.CFAIL, 0)

	// Initialize moves + obtain correct answer
	movelist := []int{ int(paxosrpc.Left), int(paxosrpc.Right), int(paxosrpc.Left) }
	initial, score, _, _ := cli.GetGameState()
	b, sc, o, w := util.CalculateGameState(initial, score, movelist)

	// Step 4: Test
	for _, m := range(movelist) {
		cli.InputMove(m)
		time.Sleep(1 * time.Second)
	}
	cb, cs, co, cw := cli.GetGameState()

	newComparison := &tests.Gamecomparison {
		b,
		sc,
		o,
		w,
		cb,
		cs,
		co,
		cw,
	}
	if !newComparison.CompareGame() {
		LOGV.Println("PHAIL: STATES ARE NOT CORRECTZ")
	} else {
		LOGV.Println("PASS: SIMPLE TEST HAS PASSED")
	}
	cli.Close()
	os.Exit(0)
}
