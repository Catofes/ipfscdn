package manager

import (
	"flag"
	"time"

	"github.com/Catofes/ipfscdn/manager/sql"
)

func int2time(i int) time.Duration {
	return time.Duration(i) * time.Millisecond
}

func main() {
	cf := flag.String("c", "test.json", "config file path.")
	flag.Parse()
	c = (&config{}).load(*cf)
	sql.Init(c.PSQL)
}
