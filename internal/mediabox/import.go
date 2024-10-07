package mediabox

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sync"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"
)

func ImportOriginals(userName string, conf *config.Config) error {
	user := entity.FindUser(userName)

	logger.Debugf("import at albums, user %s, album: %v", userName, user.Albums)
	for _, album := range user.Albums {
		photos, err := entity.FindUnimportedPhotosByAlbum(album.ID)
		if err != nil {
			continue
		}

		logger.Debugf("import at photos, %v", photos)
		for _, photo := range photos {
			filePath := filepath.Join(conf.StoragePath, "uploads", userName, album.UUID, photo.FileName)

			originalsPath := filepath.Join(conf.StoragePath, "originals", userName, album.UUID)
			if _, err := os.Stat(originalsPath); os.IsNotExist(err) {
				err := os.MkdirAll(originalsPath, os.ModePerm)
				if err != nil {
					continue
				}
			}

			destPath := filepath.Join(originalsPath, photo.FileName)
			logger.Debugf("rename %s to %s", filePath, destPath)
			if err := os.Rename(filePath, destPath); err != nil {
				logger.Debugf("rename failed, err: %s", err)
				continue
			}

			logger.Debugf("set photo %d to imported", photo.ID)
			if err := photo.SetIsImported(true); err != nil {
				logger.Debugf("set is imported failed, err : %s", err)
				continue
			}

			thumbnailsPath := filepath.Join(conf.StoragePath, "thumbnails", userName, album.UUID)
			if _, err := os.Stat(thumbnailsPath); os.IsNotExist(err) {
				err := os.MkdirAll(thumbnailsPath, os.ModePerm)
				if err != nil {
					continue
				}
			}

			thumbnailPath := filepath.Join(thumbnailsPath, fmt.Sprintf("%s.jpg", photo.FileName))
			if err = CreateThumbnail(destPath, thumbnailPath); err != nil {
				logger.Debugf("create thumbnail failed, err: %s", err)
				continue
			}
		}
	}

	return nil
}

func ImportOriginalsFromWebDAV(userName string, conf *config.Config) error {
	fileChan := make(chan FileInfoWrapper)
	webDAVDir := filepath.Join(conf.StoragePath, "user", userName)

	user := entity.FindUser(userName)

	album, err := entity.GetDefaultAlbum(userName)
	if err != nil {
		return err
	}

	go walkDir(webDAVDir, fileChan)

	var wg sync.WaitGroup
	for fileInfo := range fileChan {
		wg.Add(1)

		go func(fileInfo FileInfoWrapper) {
			defer wg.Done()

			contentType := mime.TypeByExtension(filepath.Ext(fileInfo.Path))
			uniqueFileName := GenerateUniqueFilename(fileInfo.Info.Name())

			originalsPath := filepath.Join(conf.StoragePath, "originals", userName, album.UUID)
			if _, err := os.Stat(originalsPath); os.IsNotExist(err) {
				err := os.MkdirAll(originalsPath, os.ModePerm)
				if err != nil {
					return
				}
			}

			photo := &entity.Photo{
				Name:     filepath.Base(fileInfo.Info.Name()),
				FileName: filepath.Base(uniqueFileName),
				FileSize: fileInfo.Info.Size(),
				FileType: contentType,
				AlbumID:  album.ID,
				UserID:   user.ID,
			}

			destPath := filepath.Join(originalsPath, photo.FileName)

			logger.Debugf("rename %s to %s", fileInfo.Path, destPath)
			if err := os.Rename(fileInfo.Path, destPath); err != nil {
				logger.Debugf("rename failed, err: %s", err)
				return
			}

			pho, err := photo.Create(album.ID)
			if err != nil {
				return
			}

			if err := pho.SetIsImported(true); err != nil {
				logger.Debugf("set is imported failed, err : %s", err)
				return
			}

			thumbnailsPath := filepath.Join(conf.StoragePath, "thumbnails", userName, album.UUID)
			if _, err := os.Stat(thumbnailsPath); os.IsNotExist(err) {
				err := os.MkdirAll(thumbnailsPath, os.ModePerm)
				if err != nil {
					return
				}
			}

			thumbnailPath := filepath.Join(thumbnailsPath, fmt.Sprintf("%s.jpg", photo.FileName))
			if err = CreateThumbnail(destPath, thumbnailPath); err != nil {
				logger.Debugf("create thumbnail failed, err: %s", err)
				return
			}
		}(fileInfo)
	}

	wg.Wait()
	return nil
}

type FileInfoWrapper struct {
	Path string
	Info os.FileInfo
}

func walkDir(dirPath string, fileChan chan<- FileInfoWrapper) {
	defer close(fileChan)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileChan <- FileInfoWrapper{Path: path, Info: info}
		}
		return nil
	})

	if err != nil {
		fmt.Println("Failed to walk the directory:", err)
	}
}
