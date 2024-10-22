package mediabox

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"sync"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"
	"github.com/skyline93/mediabox/internal/fs"
	"github.com/skyline93/mediabox/internal/mediabox/image"
)

type ImportJob struct {
	conf *config.Config

	user  *entity.User
	album *entity.Album
	photo *entity.Photo
}

func (i *ImportJob) Run() error {
	uPath := filepath.Join(i.conf.StoragePath, "uploads", i.user.UUID, i.album.UUID, i.photo.FileName)
	uploadPath := filepath.Join(i.conf.StoragePath, "uploads", i.user.UUID, i.album.UUID, i.photo.Name)

	logger.Debugf("rename %s to %s", uPath, uploadPath)
	if err := os.Rename(uPath, uploadPath); err != nil {
		logger.Debugf("rename failed, err: %s", err)
		return err
	}

	originalsPath := filepath.Join(i.conf.StoragePath, "originals", i.user.UUID, i.album.UUID)
	if _, err := os.Stat(originalsPath); os.IsNotExist(err) {
		err := os.MkdirAll(originalsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	destPath := filepath.Join(originalsPath, i.photo.FileName)

	thumbnailsPath := filepath.Join(i.conf.StoragePath, "thumbnails", i.user.UUID, i.album.UUID)
	if _, err := os.Stat(thumbnailsPath); os.IsNotExist(err) {
		err := os.MkdirAll(thumbnailsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	thumbnailPath := filepath.Join(thumbnailsPath, fmt.Sprintf("%s.jpg", i.photo.FileName))
	if err := image.CreateThumbnail(uploadPath, thumbnailPath, fs.IsRAWData(i.photo.Ext)); err != nil {
		logger.Debugf("create thumbnail failed, err: %s", err)
		return err
	}

	logger.Debugf("rename %s to %s", uploadPath, destPath)
	if err := os.Rename(uploadPath, destPath); err != nil {
		logger.Debugf("rename failed, err: %s", err)
		return err
	}

	logger.Debugf("set photo %d to imported", i.photo.ID)
	if err := i.photo.SetIsImported(true); err != nil {
		logger.Debugf("set is imported failed, err : %s", err)
		return err
	}

	return nil
}

func ImportOriginalsNew(userName string, conf *config.Config) {
	user := entity.FindUser(userName)

	logger.Debugf("import at albums, user %s, album: %v", userName, user.Albums)
	for _, album := range user.Albums {
		photos, err := entity.FindUnimportedPhotosByAlbum(album.ID)
		if err != nil {
			continue
		}

		logger.Debugf("import at photos, %v", photos)
		for _, photo := range photos {
			importer := ImportJob{conf: conf, user: user, album: &album, photo: &photo}

			logger.Infof("submit importjob, photo: %d", photo.ID)
			Pool.Submit(&importer)
		}
	}
}

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
			uPath := filepath.Join(conf.StoragePath, "uploads", user.UUID, album.UUID, photo.FileName)
			uploadPath := filepath.Join(conf.StoragePath, "uploads", user.UUID, album.UUID, photo.Name)

			logger.Debugf("rename %s to %s", uPath, uploadPath)
			if err := os.Rename(uPath, uploadPath); err != nil {
				logger.Debugf("rename failed, err: %s", err)
				continue
			}

			originalsPath := filepath.Join(conf.StoragePath, "originals", user.UUID, album.UUID)
			if _, err := os.Stat(originalsPath); os.IsNotExist(err) {
				err := os.MkdirAll(originalsPath, os.ModePerm)
				if err != nil {
					continue
				}
			}

			destPath := filepath.Join(originalsPath, photo.FileName)

			thumbnailsPath := filepath.Join(conf.StoragePath, "thumbnails", user.UUID, album.UUID)
			if _, err := os.Stat(thumbnailsPath); os.IsNotExist(err) {
				err := os.MkdirAll(thumbnailsPath, os.ModePerm)
				if err != nil {
					continue
				}
			}

			thumbnailPath := filepath.Join(thumbnailsPath, fmt.Sprintf("%s.jpg", photo.FileName))
			if err = image.CreateThumbnail(uploadPath, thumbnailPath, fs.IsRAWData(photo.Ext)); err != nil {
				logger.Debugf("create thumbnail failed, err: %s", err)
				continue
			}

			logger.Debugf("rename %s to %s", uploadPath, destPath)
			if err := os.Rename(uploadPath, destPath); err != nil {
				logger.Debugf("rename failed, err: %s", err)
				continue
			}

			logger.Debugf("set photo %d to imported", photo.ID)
			if err := photo.SetIsImported(true); err != nil {
				logger.Debugf("set is imported failed, err : %s", err)
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

			originalsPath := filepath.Join(conf.StoragePath, "originals", user.UUID, album.UUID)
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

			pho, err := photo.Create(album.ID)
			if err != nil {
				return
			}

			thumbnailsPath := filepath.Join(conf.StoragePath, "thumbnails", user.UUID, album.UUID)
			if _, err := os.Stat(thumbnailsPath); os.IsNotExist(err) {
				err := os.MkdirAll(thumbnailsPath, os.ModePerm)
				if err != nil {
					return
				}
			}

			thumbnailPath := filepath.Join(thumbnailsPath, fmt.Sprintf("%s.jpg", photo.FileName))
			if err = image.CreateThumbnail(destPath, thumbnailPath, fs.IsRAWData(pho.Ext)); err != nil {
				logger.Debugf("create thumbnail failed, err: %s", err)
				return
			}

			logger.Debugf("rename %s to %s", fileInfo.Path, destPath)
			if err := os.Rename(fileInfo.Path, destPath); err != nil {
				logger.Debugf("rename failed, err: %s", err)
				return
			}

			if err := pho.SetIsImported(true); err != nil {
				logger.Debugf("set is imported failed, err : %s", err)
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
