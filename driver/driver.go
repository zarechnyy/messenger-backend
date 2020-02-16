package driver

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func ConnectDB() *gorm.DB {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=keknavek dbname=test password=qwe123 sslmode=disable")

	if err != nil {
		println(err.Error())
		panic("error postgres failed to connect")
	}

	return db
}
