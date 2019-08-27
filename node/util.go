package node

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-resty/resty"
)

var privateIPBlocks []*net.IPNet
var rest *resty.Client

func path2ID(path string) string {
	t := strings.Split(path, "/")
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}

type nodeInfo struct {
	NodeID   string
	NodeAddr string
	IpfsPath []string
	DiskSize int
	Type     int
}

func int2time(i int) time.Duration {
	return time.Duration(i) * time.Millisecond
}

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
	rest = resty.New()
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}
