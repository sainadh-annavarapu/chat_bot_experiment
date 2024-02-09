package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func initDB() {
	dsn := "root:password@tcp(localhost:3306)/users_db?parseTime=true"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the database")
	}

	// db.AutoMigrate(&Teacher{}, &Student{}, &Course{}, &Enrollment{})

	fmt.Println("Gorm is Ok!")
}
