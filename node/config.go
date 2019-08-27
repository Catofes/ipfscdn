package node

import (
	"encoding/json"
	"io/ioutil"
	mlog "log"
)

type config struct {
	NodeID        string //ipfsID
	Listen        string //ListenLocalAddress
	RPCListen     string
	IpfsAPI       string //IpfsAPIAddress
	IpfsGateway   string //IpfsGateway
	Key           string //SecretKey
	ThreadNum     int
	ManagerAddr   string //CenterServerAddress
	NodeAddr      string //ListenPublicAddress
	Verbose       bool
	GraphiteAddr  string
	GraphitePath  string
	GraphiteEvery int
	DiskSize      int
	Type          int //1 fullstroe, 2 cache, 3 replica
}

func (s *config) load(path string) *config {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		mlog.Fatal(err)
	}
	s.ThreadNum = 1
	s.DiskSize = 10 * 1024 * 1024 * 1024
	s.Type = 2
	err = json.Unmarshal(data, s)
	if err != nil {
		mlog.Fatal(err)
	}
	if s.NodeAddr == "" {
		s.NodeAddr = s.Listen
	}
	return s
}
