package model

import (
	"github.com/jinzhu/gorm"
)

// Survey ...
type Survey struct {
	gorm.Model
	Uid     string `gorm:"primary_key"`
	Data    string
	Answers []Answer
}