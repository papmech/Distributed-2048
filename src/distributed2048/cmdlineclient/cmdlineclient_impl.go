package cmdlineclient

import (
	"time"
	"net/http"
	"distributed2048/lib2048"
	"distributed2048/centralserver"
	"distributed2048/rpc/paxosrpc"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"distributed2048/util"
	"os"
)

type cclient struct {
	conn *websocket.Conn
	game lib2048.Game2048
	movelist []paxosrpc.Direction
	quitchan chan int
	stoplisten chan int
}

var LOGV = util.NewLogger(true, "CMDLINECLIENT", os.Stdout)
var LOGE = util.NewLogger(true, "CMDLINECLIENT", os.Stderr)

func NewCClient(cservAddr string, interval int) (Cclient, error) {
	// Get server addr from central server
	isReady := false
	hostport := ""
	for (!isReady) {
		resp, err := http.Get(cservAddr)
		if err != nil {
			LOGV.Println("Could not connect to central server.")
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			LOGV.Println("Your mother phat")
			return nil, err
		}
		LOGV.Println("received data from cserv")
		unpacked := &centralserver.HttpReply{}
		err = json.Unmarshal(data, &unpacked)
		if err != nil {
			LOGV.Println("Your mother phat")
			return nil, err
		}
		isReady = unpacked.Status == "OK"
		if isReady {
			hostport = unpacked.Hostport
		}
		time.Sleep(1000 * time.Millisecond)
	}

	// Connect to the server
	origin := "http://localhost/"
	url := "ws://" + hostport + "/abc"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		LOGV.Println("Could not open websocket connection to server")
		return nil, err
	}
	game := lib2048.NewGame2048()
	cc := &cclient {
		ws,
		game,
		make([]paxosrpc.Direction, 0),
		make(chan int),
		make(chan int),
	}
	// Fire the ticker
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	go cc.tickHandler(ticker)
	go cc.Listen()
	return cc, nil
}

// Ticker Function
func (c *cclient) tickHandler(ticker *time.Ticker) {
	defer LOGV.Println("client has stopped ticking.")
	for {
		select {
		case <- ticker.C:
			length := len(c.movelist)
			if length > 0 {
				LOGV.Println("movelist is length " + strconv.Itoa(length))
				var translatedMove int
				switch c.movelist[length - 1] {
				case paxosrpc.Up:
					translatedMove = 0
				case paxosrpc.Right:
					translatedMove = 1
				case paxosrpc.Down:
					translatedMove = 2
				case paxosrpc.Left:
					translatedMove = 3
				}
				move := util.ClientMove{translatedMove}
				websocket.JSON.Send(c.conn, move)
				c.movelist = c.movelist[0:0]
			}
		case <- c.quitchan:
			ticker.Stop()
			return
		}
	}
}

func (c *cclient) Close() {
	LOGV.Println("closing...")
	c.quitchan <- 1
	c.stoplisten <- 1
	c.conn.Close()
}

func (c *cclient) InputMove(move paxosrpc.Direction) {
	LOGV.Println("client has input move: " + strconv.Itoa(int(move)))
	c.movelist = append(c.movelist, move)
}

func (c *cclient) GetGameState() (lib2048.Grid, int, bool, bool) {
	return c.game.GetBoard(), c.game.GetScore(), c.game.IsGameOver(), c.game.IsGameWon()
}

func (c *cclient) Listen() {
	LOGV.Println("Listening to messages from the server")
	defer LOGV.Println("Stopped listening to server")
	for {
		select {
		case <- c.stoplisten:
			return
		default:
			var data []byte
			newState := &util.Game2048State{}
			websocket.Message.Receive(c.conn, &data)
			err := json.Unmarshal(data, newState)
			if err != nil {
				LOGE.Println(err)
			}
			LOGV.Print("Trying to set the board to: ")
			LOGV.Print(newState.Grid)
			c.game.SetGameState(newState.Grid, newState.Score)
		}
	}
}
