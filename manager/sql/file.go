package sql

import (
	"github.com/jinzhu/gorm"
)

//File struct store everything about a file.
type File struct {
	gorm.Model
	Hash        string
	Name        string
	Path        string
	ParentID    uint
	Size        int64
	ContentType string
	Replica     int
	Nodes       []Node `gorm:"many2many:file_nodes;"`
}
