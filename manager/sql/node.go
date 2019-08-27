package sql

import (
	"github.com/jinzhu/gorm"
)

const (
	//CNodeTypeFullStore define that node should pin every files.
	CNodeTypeFullStore = iota + 1
	//CNodeTypeCache only cache files.
	CNodeTypeCache
	//CNodeTypeStore store files that has more than one replica.
	CNodeTypeStore
)

//Node struct in database store everything that manager need known.
type Node struct {
	gorm.Model
	Hash     string `gorm:"not_null;unique;index"`
	Address  string
	Online   bool
	DiskSize int64
	Type     int
}
