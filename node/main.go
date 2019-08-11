package node

import (
	"flag"
	"strings"
	"time"
)

func int2time(i int) time.Duration {
	return time.Duration(i) * time.Millisecond
}

//Main function is the entrance.
func Main() {
	cf := flag.String("c", "test.json", "config file path.")
	flag.Parse()
	c := (&config{}).load(*cf)
	log = (&logManager{config: *c}).init()
	n := &node{config: *c}
	n.init().loop()
}

func path2ID(path string) string {
	t := strings.Split(path, "/")
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}
