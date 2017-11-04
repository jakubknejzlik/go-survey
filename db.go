package main

import (
	"fmt"
	"net/url"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/jakubknejzlik/go-survey/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// NewDB ...
func NewDB(urlString string) (*gorm.DB, error) {

	URL, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}

	if URL == nil {
		panic("SMTP url not provided")
	}

	if URL.Scheme == "sqlite3" {
		urlString = URL.Path
	}

	fmt.Println("connecting to", URL.Scheme)
	db, err := gorm.Open(URL.Scheme, urlString)
	db.LogMode(true)

	db.AutoMigrate(&model.Answer{})
	db.AutoMigrate(&model.Survey{})
	return db, err
}
