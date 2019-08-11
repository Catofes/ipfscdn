package node

import (
	"context"
	"testing"

	shell "github.com/ipfs/go-ipfs-api"
)

func TestIPFS(t *testing.T) {
	p, e := shell.NewLocalShell().SwarmPeers(context.Background())
	if e != nil {
		t.Error(e)
	}
	t.Log(p)
}
