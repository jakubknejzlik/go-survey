package model

import "time"

// Answer ...
type Answer struct {
	Uid       string `gorm:"primary_key"`
	Data      string
	Survey    Survey `sql:"-"` // Ignore this field
	SurveyUid string `sql:"index"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
