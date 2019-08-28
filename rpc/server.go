package rnode

const MasterServiceName = "com.catofes.ipfscdn.MasterService"

type MasterService = interface {
	Sync(node NodeInfo, reply *[]NodeInfo) error
}
