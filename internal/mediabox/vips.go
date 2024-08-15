package mediabox

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/sirupsen/logrus"
)

const (
	MiB               = 1024 * 1024
	GiB               = 1024 * MiB
	DefaultCacheMem   = 128 * MiB
	DefaultCacheSize  = 128
	DefaultCacheFiles = 16
	DefaultWorkers    = 1
)

var (
	MaxCacheMem   = DefaultCacheMem
	MaxCacheSize  = DefaultCacheSize
	MaxCacheFiles = DefaultCacheFiles
	NumWorkers    = DefaultWorkers
)

var (
	vipsStarted = false
	vipsStart   = sync.Once{}
)

// VipsInit initializes libvips by checking its version and loading the ICC profiles once.
func VipsInit() {
	vipsStart.Do(vipsInit)
}

// VipsShutdown shuts down libvips and removes temporary files.
func VipsShutdown() {
	if vipsStarted {
		vipsStarted = false
		vipsStart = sync.Once{}
		vips.Shutdown()
	}
}

// vipsInit calls vips.Startup() to initialize libvips.
func vipsInit() {
	if vipsStarted {
		logger.Warnf("vips: already initialized - you may have found a bug")
		return
	}

	vipsStarted = true

	// Configure logging.
	vips.LoggingSettings(func(domain string, level vips.LogLevel, msg string) {
		switch level {
		case vips.LogLevelError, vips.LogLevelCritical:
			logger.Errorf("%s: %s", strings.TrimSpace(strings.ToLower(domain)), msg)
		case vips.LogLevelWarning:
			logger.Debugf("%s: %s", strings.TrimSpace(strings.ToLower(domain)), msg)
		default:
			logger.Tracef("%s: %s", strings.TrimSpace(strings.ToLower(domain)), msg)
		}
	}, vipsLogLevel())

	// Start libvips.
	logger.Info("startup vips")
	vips.Startup(vipsConfig())
}

// vipsConfig provides the config for initializing libvips.
func vipsConfig() *vips.Config {
	return &vips.Config{
		MaxCacheMem:      MaxCacheMem,
		MaxCacheSize:     MaxCacheSize,
		MaxCacheFiles:    MaxCacheFiles,
		ConcurrencyLevel: NumWorkers,
		ReportLeaks:      false,
		CacheTrace:       false,
		CollectStats:     false,
	}
}

// vipsLogLevel provides the libvips equivalent of the current log level.
func vipsLogLevel() vips.LogLevel {
	switch logger.GetLevel() {
	case logrus.DebugLevel:
		return vips.LogLevelWarning
	case logrus.TraceLevel:
		return vips.LogLevelDebug
	default:
		return vips.LogLevelError
	}
}

func CreateThumbnail(source string, target string) error {
	image, err := vips.NewImageFromFile(source)
	if err != nil {
		return err
	}

	if err = image.AutoRotate(); err != nil {
		return err
	}

	ep := vips.NewDefaultJPEGExportParams()
	imageBytes, _, err := image.Export(ep)
	if err != nil {
		return err
	}

	return os.WriteFile(target, imageBytes, 0644)
}

// CreateThumbnailsInDirectory 遍历目录下所有图片并创建缩略图
func CreateThumbnailsInDirectory(srcDir string, destDir string) error {
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
			if err := CreateThumbnail(path, fmt.Sprintf("%s.jpg", destPath)); err != nil {
				return fmt.Errorf("failed to create thumbnail for %s: %w", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking the path %s: %w", srcDir, err)
	}

	return nil
}
