package mediabox

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"
)

func CleanUpExpiredPhotos(conf *config.Config) error {
	logger.Infof("starting cleanup expired photos")
	expiredPhotos, err := entity.FindExpiredPhotos()
	if err != nil {
		return err
	}

	logger.Debugf("expiredPhotos: %v", expiredPhotos)
	for _, photo := range expiredPhotos {
		user := photo.User
		album := photo.Album

		deletePhotoFile(user.Name, album.Name, photo.FileName, conf)

		err = entity.DeletePhotoRecord(&photo)
		if err != nil {
			logger.Errorf("failed to delete photo record from database, error: %v", err)
			return fmt.Errorf("failed to delete photo record from database: %v", err)
		}
	}

	return nil
}

func deletePhotoFile(userName, albumName, fileName string, conf *config.Config) {
	files := []string{
		filepath.Join(conf.StoragePath, "thumbnails", userName, albumName, fmt.Sprintf("%s.jpg", fileName)),
		filepath.Join(conf.StoragePath, "originals", userName, albumName, fileName),
	}

	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			logger.Errorf("failed to delete photo file: %s, error: %v", file, err)
			continue
		}
	}
}
