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
}

func (s *node) init() *node {
	s.ctx, s.done = context.WithCancel(context.Background())
	return s
}

func (s *node) ping() bool {
	response, err := http.Get(s.WebAddress + "/generate_204")
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

func (s *node) get(file *sql.File) (*http.Response, error) {
	response, err := http.Get(fmt.Sprintf("%s/file/%s", s.WebAddress, file.Hash))
	if err != nil && response.StatusCode == 200 {
		return response, nil
	}
	log.Warning("")
	if err != nil {
		return nil, err
	}
	return nil, errors.New("bad response ")
}
