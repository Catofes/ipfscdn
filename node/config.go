package node

import (
	"encoding/json"
	"io/ioutil"
	mlog "log"
)

type config struct {
	NodeID        string
	Listen        string
	IpfsAddr      string
	IpfsGateway   string
	Key           string
	ThreadNum     int
	ManagerAddr   string
	NodeAddr      string
	Verbose       bool
	GraphiteAddr  string
	GraphitePath  string
	GraphiteEvery int
}

func (s *config) load(path string) *config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		mlog.Fatal(err)
	}
	s.ThreadNum = 1
	err = json.Unmarshal(data, s)
	if err != nil {
		mlog.Fatal(err)
	}
	if s.NodeAddr == "" {
		s.NodeAddr = s.Listen
	}
	return s
}
