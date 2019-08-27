package node

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
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
	response, err := rest.R().
		SetAuthToken(s.node.config.Key).
		SetBody(nodeInfo{
			NodeID:   s.node.NodeID,
			NodeAddr: s.node.NodeAddr,
			IpfsPath: s.node.ipfsAddr}).
		Put(fmt.Sprintf("%s/node/%s", s.webAddress, s.node.NodeID))

	if err != nil && response.StatusCode() == 200 {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("bad response ")
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
		time.Sleep(30 * time.Second)
	}
}
