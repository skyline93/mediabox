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
