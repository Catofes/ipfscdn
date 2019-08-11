package node

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type peer struct {
	ID         string
	path       []string
	online     bool
	lastOnline time.Time
	mutex      sync.Mutex
	node       *node
}

func (s *peer) init(node *node) *peer {
	s.path = make([]string, 0)
	s.node = node
	return s
}

func (s *peer) addPath(p []string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.path = p
}

func (s *peer) check() bool {
	peers, err := s.node.ipfs.SwarmPeers(context.Background())
	if err != nil {
		log.Warning("Get swarm peers failed: %s.", err.Error)
		return false
	}
	for _, p := range peers.Peers {
		if p.Peer == s.ID {
			return true
		}
	}
	return false
}

func (s *peer) connect() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	p := rand.Intn(len(s.path))
	err := s.node.ipfs.SwarmConnect(context.Background(), s.path[p])
	return err
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
