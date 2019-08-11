package manager

import (
	"encoding/json"
	"io/ioutil"
	mlog "log"
)

var c *config

type config struct {
	PSQL          string
	Listen        string
	CheckEvery    int
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
	s.CheckEvery = 6000
	s.GraphiteEvery = 1000
	s.Verbose = false
	err = json.Unmarshal(data, s)
	if err != nil {
		mlog.Fatal(err)
	}
	return s
}
