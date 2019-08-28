package rnode

import (
	"net/rpc"

	shell "github.com/ipfs/go-ipfs-api"
)

type NodeInfo struct {
	NodeID   string
	NodeAddr string
	IpfsPath []string
	DiskSize int
	Type     int
}

type FileLists = map[string]shell.PinInfo

const NodeServiceName = "com.catofes.ipfscdn.NodeService"

type NodeService = interface {
	Pin(fileHash string, reply *bool) error
	UnPin(fileHash string, reply *bool) error
	Connect(node NodeInfo, _ *bool) error
	List(_ bool, lists *FileLists) error
	Ping(_ bool, _ *bool) error
}

func RegisterNodeService(svc NodeService) error {
	return rpc.RegisterName(NodeServiceName, svc)
}

type NodeServiceClient struct {
	*rpc.Client
}

var _ NodeService = (*NodeServiceClient)(nil)

func DialNodeService(network, address string) (*NodeServiceClient, error) {
	c, err := rpc.Dial(network, address)
	if err != nil {

		return nil, err
	}
	return &NodeServiceClient{Client: c}, nil
}

func (p *NodeServiceClient) Pin(fileHash string, reply *bool) error {
	return p.Client.Call(NodeServiceName+".Pin", fileHash, reply)
}

func (p *NodeServiceClient) UnPin(fileHash string, reply *bool) error {
	return p.Client.Call(NodeServiceName+".UnPin", fileHash, reply)
}

func (p *NodeServiceClient) Connect(node NodeInfo, _ *bool) error {
	return p.Client.Call(NodeServiceName+".Connect", node, true)
}

func (p *NodeServiceClient) List(_ bool, reply *FileLists) error {
	return p.Client.Call(NodeServiceName+".List", true, reply)
}

func (p *NodeServiceClient) Ping(_ bool, _ *bool) error {
	return p.Client.Call(NodeServiceName+".Ping", true, true)
}
