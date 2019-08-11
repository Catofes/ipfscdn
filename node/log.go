package node

import (
	"net"
	"os"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/op/go-logging"
	"github.com/rcrowley/go-metrics"
)

var log *logManager

type logManager struct {
	config
	*logging.Logger
	metrics.Registry
	id string
}

func (s *logManager) init() *logManager {
	s.Logger = logging.MustGetLogger("example")
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{color}%{time:0102 15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(backendFormatter)
	if s.config.Verbose {
		backendLeveled.SetLevel(logging.DEBUG, "")
	} else {
		backendLeveled.SetLevel(logging.WARNING, "")
	}
	logging.SetBackend(backendLeveled)

	if s.GraphitePath != "" {
		addr, err := net.ResolveTCPAddr("tcp", s.config.GraphiteAddr)
		if err != nil {
			s.Warning("graphite address wrong")
			return s
		}
		s.Registry = metrics.DefaultRegistry
		go graphite.Graphite(s.Registry, int2time(s.config.GraphiteEvery), s.config.GraphitePath, addr)
	}

	return s
}

func (s *logManager) Mark(path string, value int) {
	t := metrics.GetOrRegisterGauge("node."+s.NodeID+path, log.Registry)
	t.Update(int64(value))
}
