package node

import (
	"context"
	"math/rand"
	"sync"
	"time"

	rnode "github.com/Catofes/ipfscdn/rpc"
)

type peer struct {
	NodeID     string
	path       []string
	webAddress string
	online     bool
	lastOnline time.Time
	mutex      sync.Mutex
	node       *Node
}

func (s *peer) init(node *Node) *peer {
	s.path = make([]string, 0)
	s.node = node
	return s
}

func (s *peer) addPath(p []string, a string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.path = p
	s.webAddress = a
}

func (s *peer) check() bool {
	peers, err := s.node.ipfs.SwarmPeers(context.Background())
	if err != nil {
		log.Warning("Get swarm peers failed: %s.", err.Error)
		return false
	}
	for _, p := range peers.Peers {
		if p.Peer == s.NodeID {
			return true
		}
	}
	return false
}

func (s *peer) connectWakeUp() error {
	client, err := rnode.DialNodeService("tcp", "localhost:1234")
	if err != nil {
		log.Debugf("[%s] RPC dail failed, %s.", s.NodeID, err.Error())
		return err
	}
	err = client.Connect(rnode.NodeInfo{
		NodeID:   s.node.NodeID,
		NodeAddr: s.node.NodeAddr,
		IpfsPath: s.node.ipfsAddr,
	}, nil)
	if err != nil {
		log.Debugf("[%s] RPC connect failed, %s.", s.NodeID, err.Error())
		return err
	}
	return nil
}

func (s *peer) connect() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	p := rand.Intn(len(s.path))
	err := s.node.ipfs.SwarmConnect(context.Background(), s.path[p])
	if err != nil {
		return err
	}
	return s.connectWakeUp()
}

func (s *peer) loop() {
	for {
		ok := s.check()
		if ok {
			s.lastOnline = time.Now()
			s.online = true
		} else {
			s.online = false
			s.connect()
		}
		time.Sleep(1 * time.Minute)
	}
}
