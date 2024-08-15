package entity

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name     string  `gorm:"size:200;index;"`
	Password string  `gorm:"size:500"`
	Role     string  `gorm:"size:60"`
	Albums   []Album `gorm:"many2many:users_albums;" yaml:"-"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) InvalidPassword(s string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(s)) == nil
}

func CreateUser(name, password, role string) error {
	if err := Db().Create(&User{Name: name, Password: password, Role: role}).Error; err != nil {
		return err
	}

	return nil
}

func FindUser(name string) *User {
	result := User{}
	if err := Db().Where("name = ?", name).Preload("Albums").First(&result).Error; err != nil {
		return nil
	}

	return &result
}

func DeleteUser(name string) error {
	if err := UnscopedDb().Delete(User{}, "name = ?", name).Error; err != nil {
		return err
	}

	return nil
}
