// A simple 2048 client running on a command line that facilitates automated testing

package cmdlineclient

type Cclient interface {
	InputMove()
	GetGameState()
	Close()
}
