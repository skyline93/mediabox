package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Album struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UUID   string  `json:"uuid"`
	Name   string  `gorm:"size:200;index;" json:"name"`
	Users  []User  `gorm:"many2many:users_albums;" json:"-"`
	Photos []Photo `gorm:"many2many:photos_albums;" json:"-"`
}

func (Album) TableName() string {
	return "albums"
}

func ExistsAlbum(name string, userName string) bool {
	var user User

	err := Db().Model(&User{}).Where("name = ?", userName).Preload("Albums").First(&user).Error
	if err != nil {
		return false
	}

	for _, album := range user.Albums {
		if album.Name == name {
			return true
		}
	}

	return false
}

func (s *Album) Create(name string, userID uint) (*Album, error) {
	album := &Album{Name: s.Name, UUID: s.UUID}

	if err := Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(album).Error; err != nil {
			return err
		}

		var user User
		if err := tx.First(&user, userID).Error; err != nil {
			logger.Debugf("err: %s", err)
			return err
		}

		if err := tx.Model(&user).Association("Albums").Append(album); err != nil {
			logger.Debugf("err: %s", err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return album, nil
}

func FindAlbumByID(id uint) (*Album, error) {
	var album Album

	if err := Db().First(&album, id).Error; err != nil {
		return nil, err
	}

	return &album, nil
}

func (s *Album) Update(name string) error {
	if err := Db().Model(&Album{}).Where("id = ?", s.ID).Update("name", name).Error; err != nil {
		return err
	}

	return nil
}

func (s *Album) Delete(id uint) error {
	if err := Db().Delete(&Album{}, id).Error; err != nil {
		return err
	}

	return nil
}

func ListAlbumsByUserID(userID uint) ([]Album, error) {
	var user User

	err := Db().Model(&User{}).Where("id = ?", userID).Preload("Albums").First(&user).Error
	if err != nil {
		return nil, err
	}

	return user.Albums, err
}

func GetDefaultAlbum(userName string) (*Album, error) {
	var user User

	err := Db().Model(&User{}).Where("name = ?", userName).Preload("Albums").First(&user).Error
	if err != nil {
		return nil, err
	}

	for _, album := range user.Albums {
		if album.Name == "default" {
			return &album, nil
		}
	}

	return nil, fmt.Errorf("default album is not exists")
}
