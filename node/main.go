package node

import (
	"flag"
)

//Main function is the entrance.
func Main() {
	cf := flag.String("c", "test.json", "config file path.")
	flag.Parse()
	c := (&config{}).load(*cf)
	log = (&logManager{config: *c}).init()
	n := &Node{config: *c}
	n.init().loop()
}
