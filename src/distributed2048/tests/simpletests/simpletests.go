package main

// ======================================================================== //
// Dis be berry simple test
// Dere be 1 central server + 1 game server + 1 client
// Der client be making 8 simple moves
// ======================================================================== //
import (
	"distributed2048/cmdlineclient"
	"distributed2048/lib2048"
	"distributed2048/util"
	"fmt"
	"os"
	"time"
)

var LOGV = util.NewLogger(false, "LIBSTORETEST", os.Stdout)

var (
	passCount int
	failCount int
)

type testFunc struct {
	name string
	f    func()
}

func processError(err error, msg string) {
	if err != nil {
		LOGV.Println(err)
		LOGV.Println(msg)
		os.Exit(1)
	}
}

func testOneCentralOneClientOneGameserv() {
	// Step 1: Boot Testing Client
	cservAddr := "http://" + util.CENTRALHOSTPOST
	cli, err := cmdlineclient.NewCClient(cservAddr, "", util.DEFAULTINTERVAL)
	processError(err, util.CFAIL)

	// Step 2: Initialize moves + obtain correct answer
	movelist := []*lib2048.Move{
		lib2048.NewMove(lib2048.Left),
		lib2048.NewMove(lib2048.Right),
		lib2048.NewMove(lib2048.Left),
		lib2048.NewMove(lib2048.Up),
		lib2048.NewMove(lib2048.Up),
		lib2048.NewMove(lib2048.Down),
		lib2048.NewMove(lib2048.Right),
		lib2048.NewMove(lib2048.Left),
	}
	desiredGame := util.CalculateGameState(cli.GetGameState(), movelist)

	// Step 3: Test
	for _, m := range movelist {
		cli.InputMove(m.Direction)
		time.Sleep(1000 * time.Millisecond) // has to be large enough so one move per paxos round
	}

	LOGV.Println("CLI's game state")
	LOGV.Println(cli.GetGameState())
	LOGV.Println("Desired game state")
	LOGV.Println(desiredGame)

	if !cli.GetGameState().Equals(desiredGame) {
		fmt.Println("PHAIL: STATES ARE NOT CORRECTZ")
		failCount++
		return
	}
	fmt.Println("PASS")
	passCount++
}

func main() {
	tests := []testFunc{
		{"testOneCentralOneClientOneGameserv", testOneCentralOneClientOneGameserv},
	}

	for _, test := range tests {
		fmt.Printf("Running %s:\n", test.name)
		test.f()
	}

	fmt.Printf("Passed (%d/%d) tests\n", passCount, passCount+failCount)
}
