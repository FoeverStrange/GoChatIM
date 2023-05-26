package models

import "gorm.io/gorm"

//群信息
type GroupBasic struct {
	gorm.Model
	Name    string
	OwnerId uint
	Icon    string
	Type    int    //多少人的群，level
	Desc    string //描述
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
