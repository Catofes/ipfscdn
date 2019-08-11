package main

import (
	"flag"

	"github.com/Catofes/ipfscdn/manager/sql"
	"github.com/jinzhu/gorm"
)

func main() {
	path := flag.String("c", "", "pg path")
	flag.Parse()
	db, err := gorm.Open("postgres", *path)
	if err != nil {
		panic("failed to connect database, err:" + err.Error())
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&sql.Node{}, &sql.File{})
}
