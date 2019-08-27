package manager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Catofes/ipfscdn/manager/sql"
	"github.com/rcrowley/go-metrics"
)

type node struct {
	config
	sql.Node
	ctx  context.Context
	done context.CancelFunc
	m    *manager
}

func (s *node) init(m *manager) *node {
	s.ctx, s.done = context.WithCancel(context.Background())
	s.m = m
	return s
}

func (s *node) ping() bool {
	response, err := http.Get(s.Address + "/generate_204")
	if err != nil {
		return false
	}
	if response.StatusCode != 204 {
		return false
	}
	return true
}

func (s *node) updateDB() {
	err := sql.Get("").Update(s.Node)
	if err != nil {
		log.Debug("[Manager:NodeStatus] update node %s failed %s.", s.ID, err)
	}
}

func (s *node) loop() {
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				s.Online = s.ping()
				t := metrics.GetOrRegisterGauge("node."+string(s.ID)+".online", log.Registry)
				if s.Online {
					t.Update(1)
				} else {
					t.Update(0)
				}
			}
			time.Sleep(int2time(s.CheckEvery))
		}
	}()
}

func (s *node)
