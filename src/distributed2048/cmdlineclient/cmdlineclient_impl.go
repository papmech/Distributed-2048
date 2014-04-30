package cmdlineclient

import (
	"code.google.com/p/go.net/websocket"
	"distributed2048/centralserver"
	"distributed2048/lib2048"
	"distributed2048/util"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type cclient struct {
	conn       *websocket.Conn
	game       lib2048.Game2048
	movelist   []lib2048.Direction
	quitchan   chan int
	stophandler chan int
	stopsender chan int
	stopreceiver chan int
	moveQueue  chan util.ClientMove
	cserv string
}

var LOGV = util.NewLogger(false, "CMDLINECLIENT", os.Stdout)
var LOGE = util.NewLogger(true, "CMDLINECLIENT", os.Stderr)

func NewCClient(cservAddr string, gameServHostPort string, interval int) (Cclient, error) {
	ws, err := doConnect(cservAddr, gameServHostPort)
	if err != nil {
		return nil, err
	}
	game := lib2048.NewGame2048()
	cc := &cclient{
		ws,
		game,
		make([]lib2048.Direction, 0),
		make(chan int),
		make(chan int),
		make(chan int),
		make(chan int),
		make(chan util.ClientMove),
		cservAddr,
	}
	// Fire the ticker
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	go cc.tickHandler(ticker)
	go cc.websocketHandler()
	return cc, nil
}

func doConnect(cservAddr string, gameServHostPort string) (*websocket.Conn, error) {
	if gameServHostPort == "" {
		// Get server addr from central server
		isReady := false
		hostport := ""
		for !isReady {
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
				gameServHostPort = hostport
				// Connect to the server
				origin := "http://localhost/"
				url := "ws://" + gameServHostPort + "/abc"
				ws, err := websocket.Dial(url, "", origin)
				if err != nil {
					LOGV.Println("Could not open websocket connection to server")
					isReady = false
				} else {
					LOGE.Println("Connection has been established with server " + gameServHostPort)
					return ws, nil
				}
			}
			time.Sleep(250 * time.Millisecond)
		}
	}

	// Connect to the server
	origin := "http://localhost/"
	url := "ws://" + gameServHostPort + "/abc"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		LOGV.Println("Could not open websocket connection to server")
		return nil, err
	} else {
		LOGE.Println("Connection has been established with server " + gameServHostPort)
		return ws, nil
	}
}

// Ticker Function
func (c *cclient) tickHandler(ticker *time.Ticker) {
	defer LOGV.Println("client has stopped ticking.")
	for {
		select {
		case <-ticker.C:
			length := len(c.movelist)
			if length > 0 {
				LOGV.Println("movelist is length " + strconv.Itoa(length))
				var translatedMove int
				switch c.movelist[length-1] {
				case lib2048.Up:
					translatedMove = 0
				case lib2048.Right:
					translatedMove = 1
				case lib2048.Down:
					translatedMove = 2
				case lib2048.Left:
					translatedMove = 3
				}
				move := util.ClientMove{translatedMove}
				c.moveQueue <- move
				c.movelist = c.movelist[0:0]
//				websocket.JSON.Send(c.conn, move)
			}
		case <-c.quitchan:
			ticker.Stop()
			close(c.quitchan)
			return
		}
	}
}

func (c *cclient) Close() {
	LOGV.Println("closing...")
	c.quitchan <- 1
	c.stophandler <- 1
	c.conn.Close()
}

func (c *cclient) InputMove(move lib2048.Direction) {
	LOGV.Println("client has input move: " + strconv.Itoa(int(move)))
	c.movelist = append(c.movelist, move)
}

func (c *cclient) GetGameState() lib2048.Game2048 {
	return c.game
}

func (c *cclient) sender() (chan<- util.ClientMove, chan error) {
	ch, errCh := make(chan util.ClientMove), make(chan error)
	go func() {
		defer LOGV.Println("sender has died")
		for {
			select {
			case <-c.stopsender:
				close(ch)
				return
			case s := <-ch:
				if err := websocket.JSON.Send(c.conn, s); err != nil {
					errCh <- err
					close(ch)
					return
				}
			}
		}
	}()
	return ch, errCh
}

func (c *cclient) receiver() (<-chan []byte, chan error) {
	ch, errCh := make(chan []byte), make(chan error)
	go func() {
		defer LOGV.Println("receiver has died")
		for {
			select {
			case <-c.stopreceiver:
				close(ch)
				return
			default:
				var s []byte
				if err := websocket.Message.Receive(c.conn, &s); err != nil {
					errCh <- err
					close(ch)
					return
				}
				ch <- s
			}
		}
	}()
	return ch, errCh
}

func (c *cclient) websocketHandler() {
	send, sendErr := c.sender()
	recv, recvErr := c.receiver()
	for {
		select {
		case s := <-c.moveQueue:
			send <- s
		case s := <-recv:
			newState := &util.Game2048State{}
			err := json.Unmarshal(s, newState)
			if err != nil {
				LOGE.Println(err)
			}
			LOGV.Print("Trying to set the board to: ")
			LOGV.Print(newState.Grid)
			c.game.SetGrid(newState.Grid)
			c.game.SetScore(newState.Score)
		case err := <-sendErr:
			LOGE.Println("Communication error with server while sending move: " + err.Error())
			c.stopreceiver <- 1
			LOGE.Println("Attempting Reconnect")
			ws, err := doConnect(c.cserv, "")
			LOGE.Println("Reconnect complete")
			if err != nil {
				LOGE.Println("Unable to reconnect to server. Shutting down..")
				go c.Close()
			}
			c.conn = ws
			send, sendErr = c.sender()
			recv, recvErr = c.receiver()
		case err := <-recvErr:
			LOGE.Println("Communication error with server while receiving state")
			c.stopsender <- 1
			LOGE.Println("Attempting Reconnect")
			ws, err := doConnect(c.cserv, "")
			LOGE.Println("Reconnect complete")

			if err != nil {
				LOGE.Println("Unable to reconnect to server. Shutting down..")
				go c.Close()
			}
			c.conn = ws
			send, sendErr = c.sender()
			recv, recvErr = c.receiver()
		case <-c.stophandler:
			<-c.stopreceiver
			<-c.stopsender
			close(c.stopreceiver)
			close(c.stopsender)
			close(c.stophandler)
			return
		}
	}
}
