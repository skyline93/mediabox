package image

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func CreateThumbnail(source, target string, isRAWData bool) error {
	if isRAWData {
		dt := NewDarktable()
		if err := dt.CreateThumbnail(source, target); err != nil {
			return fmt.Errorf("failed to create thumbnail for %s: %w", source, err)
		}
	} else {
		if err := VipsInstance.CreateThumbnail(source, target); err != nil {
			return fmt.Errorf("failed to create thumbnail for %s: %w", source, err)
		}
	}

	return nil
}

// CreateThumbnailsInDirectory 遍历目录下所有图片并创建缩略图
func CreateThumbnailsInDirectory(srcDir string, destDir string, isRAWData bool) error {
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			// 构造目标文件路径
			relPath, _ := filepath.Rel(srcDir, path)
			destPath := filepath.Join(destDir, relPath)

			// 确保目标目录存在
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return err
			}

			// 创建缩略图
			if isRAWData {
				dt := NewDarktable()
				if err := dt.CreateThumbnail(path, fmt.Sprintf("%s.jpg", destPath)); err != nil {
					return fmt.Errorf("failed to create thumbnail for %s: %w", path, err)
				}
			} else {
				if err := VipsInstance.CreateThumbnail(path, fmt.Sprintf("%s.jpg", destPath)); err != nil {
					return fmt.Errorf("failed to create thumbnail for %s: %w", path, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the path %s: %w", srcDir, err)
	}

	return nil
}
