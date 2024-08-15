package mediabox

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/skyline93/mediabox/internal/config"
)

func GenerateUniqueFilename(originalFilename string) string {
	return originalFilename + "_" + time.Now().Format("2006-01-02_15-04-05")
}

func UploadPhoto(userName string, albumName string, uniqueFileName string, file io.Reader, conf *config.Config) error {
	userDir := filepath.Join(conf.StoragePath, "uploads", userName, albumName)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		err := os.MkdirAll(userDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	fp, err := os.Create(filepath.Join(userDir, uniqueFileName))
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err := io.Copy(fp, file); err != nil {
		return err
	}

	return nil
}

// func UploadPhotoNew(userName string, albumID uint, sourceFileName string, sourceFile *os.File, conf *config.Config) (*entity.Photo, error) {
// 	album, err := entity.FindAlbumByID(uint(albumID))
// 	if err != nil {
// 		return nil, err
// 	}

// 	uniqueFileName := GenerateUniqueFilename(sourceFileName)

// 	photo := &entity.Photo{
// 		Name:     filepath.Base(sourceFileName),
// 		FileName: filepath.Base(uniqueFileName),
// 		// FileSize: sourceFile.Size,
// 		// FileType: strings.Split(file.Header.Get("Content-Type"), ";")[0],
// 	}

// 	src, err := file.Open()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer src.Close()

// 	if err := UploadPhoto(userName, album.Name, uniqueFileName, src, conf); err != nil {
// 		return nil, err
// 	}

// 	pho, err := photo.Create(uint(albumID))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return pho, nil
// }
