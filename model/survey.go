package model

import "time"

// Survey ...
type Survey struct {
	Uid     string `gorm:"primary_key"`
	Data    string
	Answers []Answer `gorm:"ForeignKey:SurveyUid;AssociationForeignKey:Uid"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}
