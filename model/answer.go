package model

import (
	"github.com/jinzhu/gorm"
)

// Survey ...
type Answer struct {
	gorm.Model
	Uid      string `gorm:"primary_key"`
	Data     string
	Survey   Survey
	SurveyID int
}
