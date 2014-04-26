package cmdlineclient

import (
	"time"
	"fmt"
	"net/http"
	"distributed2048/lib2048"
	"distributed2048/centralserver"
	"code.google.com/p/go.net/websocket"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

const (
	UP = 0
	RIGHT = 1
	DOWN = 2
	LEFT = 3
)

type cclient struct {
	conn *websocket.Conn
	game lib2048.Game2048
	movelist []int
	quitchan chan int
}

func NewCClient(cservAddr string, interval int) (Cclient, error) {
	// Get server addr from central server
	isReady := false
	hostport := ""
	for (!isReady) {
		resp, err := http.Get(cservAddr)
		if err != nil {
			fmt.Println("Could not connect to central server.")
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println("Your mother phat")
			return nil, err
		}
		fmt.Println("received data from cserv")
		unpacked := &centralserver.HttpReply{}
		err = json.Unmarshal(data, &unpacked)
		if err != nil {
			fmt.Println("Your mother phat")
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
		fmt.Println("Could not open websocket connection to server")
		return nil, err
	}
	game := lib2048.NewGame2048()
	cc := &cclient {
		ws,
		game,
		make([]int, 0),
		make(chan int),
	}
	// Fire the ticker
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	go cc.tickHandler(ticker)
	return cc, nil
}

// Ticker Function
func (c *cclient) tickHandler(ticker *time.Ticker) {
	defer fmt.Println("client has stopped ticking.")
	for {
		select {
		case <- ticker.C:
			length := len(c.movelist)
			if length > 0 {
				c.conn.Write([]byte(strconv.Itoa(c.movelist[length - 1])))
				c.movelist = c.movelist[0:0]
			}
		case <- c.quitchan:
			ticker.Stop()
			return
		}
	}
}

func (c *cclient) Close() {
	c.quitchan <- 1
}

func (c *cclient) InputMove(move int) {
	c.movelist = append(c.movelist, move)
}

func (c *cclient) GetGameState() (lib2048.Grid, int, bool, bool) {
	return c.game.GetBoard(), c.game.GetScore(), c.game.IsGameOver(), c.game.IsGameWon()
}
