package sql

import (
	"log"
	"sync"

	"github.com/jinzhu/gorm"
	//postgres
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var sql *DB
var once sync.Once

//DB struct is main database structure
type DB struct {
	address string
	*gorm.DB
}

func (s *DB) init() *DB {
	var err error
	s.DB, err = gorm.Open("postgres", s.address)
	if err != nil {
		log.Fatal(err)
	}
	return s
}

//Init func connect to remote psql db
func Init(path string) *DB {
	once.Do(func() {
		sql = (&DB{address: path}).init()
	})
	return sql
}

//Get func return the connected db
func Get(path string) *DB {
	return sql
}
