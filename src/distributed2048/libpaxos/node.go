package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
	"net/rpc"
	"sync"
)

type node struct {
	Info   paxosrpc.Node
	Client *rpc.Client
	Mutex  sync.Mutex
}

func NewNode(info paxosrpc.Node) *node {
	return &node{
		Info: info,
	}
}

func (n *node) getRPCClient() *rpc.Client {
	n.Mutex.Lock()
	if n.Client == nil {
		c, _ := rpc.DialHTTP("tcp", n.Info.HostPort)
		// if err != nil {
		// 	fmt.Println(err)
		// }
		n.Client = c
	}
	n.Mutex.Unlock()
	return n.Client
}
