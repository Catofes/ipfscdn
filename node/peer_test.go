package node

import (
	"testing"

	shell "github.com/ipfs/go-ipfs-api"
)

func TestIPFS(t *testing.T) {
	p, e := shell.NewLocalShell().ID()
	if e != nil {
		t.Error(e)
	}
	for _, v := range p.Addresses {
		t.Log(v)
	}
}
