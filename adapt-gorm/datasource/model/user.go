package model

type User struct {
	ID   uint
	Name string
}

type UserRole struct {
	ID     uint
	UserId uint `gorm:"column:user_id"`
	RoleId uint `gorm:"column:role_id"`
}
