package model

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model

	Name string `json:"name,omitempty" form:"name"`
}

type Job struct {
	gorm.Model
	Name        string `json:"name,omitempty" form:"name"`
	Description string `json:"description,omitempty" form:"description"`
}
