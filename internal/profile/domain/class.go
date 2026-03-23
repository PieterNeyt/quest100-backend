package domain

import "github.com/google/uuid"

type Course struct {
	Id      uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name    string
	Classes []*Class `gorm:"foreignKey:CourseId"`
}

type Class struct {
	Id       uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name     string
	CourseId uuid.UUID `gorm:"type:uuid"`
}
