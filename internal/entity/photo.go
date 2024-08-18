package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Photo struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name       string `gorm:"size:200;index;" json:"name"`
	FileName   string `gorm:"size:300;index;" json:"file_name"`
	FileHash   string `gorm:"size:128;index" json:"file_hash"`
	FileSize   int64  `json:"file_size"`
	FileType   string `gorm:"size:16" json:"file_type"`
	IsImported bool   `gorm:"default:false" json:"is_imported"`

	UserID  uint    `json:"user_id"`
	AlbumID uint    `json:"album_id"`
	Albums  []Album `gorm:"many2many:photos_albums;" json:"-"`

	Link string `gorm:"-" json:"link"`
}

func (Photo) TableName() string {
	return "photos"
}

func (s *Photo) Create(albumID uint) (*Photo, error) {
	photo := &Photo{
		Name:       s.Name,
		FileName:   s.FileName,
		FileHash:   s.FileHash,
		FileSize:   s.FileSize,
		FileType:   s.FileType,
		IsImported: false,
		AlbumID:    s.AlbumID,
		UserID:     s.UserID,
	}

	if err := Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(photo).Error; err != nil {
			return err
		}

		var album Album
		if err := tx.First(&album, albumID).Error; err != nil {
			logger.Debugf("err: %s", err)
			return err
		}

		if err := tx.Model(&album).Association("Photos").Append(photo); err != nil {
			logger.Debugf("err: %s", err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return photo, nil
}

func FindUnimportedPhotosByAlbum(id uint) ([]Photo, error) {
	var album Album

	if err := Db().Preload("Photos", "is_imported = ?", false).First(&album, id).Error; err != nil {
		return nil, err
	}

	return album.Photos, nil
}

func (s *Photo) SetIsImported(isImported bool) error {
	if err := Db().Model(&Photo{}).Where("id = ?", s.ID).Update("is_imported", isImported).Error; err != nil {
		return err
	}

	return nil
}

func ListPhotos(userName string, albumID uint, page, pageSize int) ([]Photo, int, int, error) {
	var user User

	err := Db().Model(&User{}).Where("name = ?", userName).Preload("Albums").First(&user).Error
	if err != nil {
		return nil, 0, 0, err
	}

	for _, album := range user.Albums {
		if album.ID == albumID {
			var alb Album
			var totalPhotos int64

			err := Db().Model(&Photo{}).Where("album_id = ?", albumID).Count(&totalPhotos).Error
			if err != nil {
				return nil, 0, 0, err
			}

			totalPages := int((totalPhotos + int64(pageSize) - 1) / int64(pageSize))

			offset := (page - 1) * pageSize
			err = Db().Model(&Album{}).Where("id = ?", albumID).
				Preload("Photos", func(db *gorm.DB) *gorm.DB {
					return db.Offset(offset).Limit(pageSize)
				}).
				First(&alb).Error

			if err != nil {
				return nil, 0, 0, err
			}

			return alb.Photos, int(totalPhotos), totalPages, nil
		}
	}

	return nil, 0, 0, fmt.Errorf("album does not exist")
}
