package libpaxos

import (
	"net/rpc"
	"sync"
)

type node struct {
	Info   PaxosNode
	Client *rpc.Client
	Mutex  sync.Mutex
}

func (n *node) getRPCClient() *rpc.Client {
	n.Mutex.Lock()
	if n.Client == nil {
		c, _ := rpc.DialHTTP("tcp", node.Info.HostPort)
		n.Client = c
	}
	n.Mutex.Unlock()
	return n.Client
}
