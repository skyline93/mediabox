package image

import (
	"errors"
	"os"
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

var VipsInstance *Vips

func init() {
	VipsInstance = NewVips()
}

type Vips struct {
	started bool
	start   sync.Once
}

func NewVips() *Vips {
	return &Vips{
		started: false,
		start:   sync.Once{},
	}
}

func (v *Vips) Init() {
	v.start.Do(func() {
		if v.started {
			logger.Warnf("vips: already initialized - you may have found a bug")
			return
		}

		v.started = true

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
		}, v.vipsLogLevel())

		// Start libvips.
		logger.Info("startup vips")
		vips.Startup(v.defaultConfig())
	})
}

// VipsShutdown shuts down libvips and removes temporary files.
func (v *Vips) Shutdown() {
	if v.started {
		v.started = false
		v.start = sync.Once{}
		vips.Shutdown()
	}
}

// vipsConfig provides the config for initializing libvips.
func (v *Vips) defaultConfig() *vips.Config {
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
func (v *Vips) vipsLogLevel() vips.LogLevel {
	switch logger.GetLevel() {
	case logrus.DebugLevel:
		return vips.LogLevelWarning
	case logrus.TraceLevel:
		return vips.LogLevelDebug
	default:
		return vips.LogLevelError
	}
}

var ErrMissingDimensions = errors.New("either width or height must be specified")

func (v *Vips) CreateThumbnail(source, target string) error {
	width := DefaultThumbnailWidth
	height := DefaultThumbnailHeight

	image, err := vips.NewImageFromFile(source)
	if err != nil {
		return err
	}

	if err = image.AutoRotate(); err != nil {
		return err
	}

	// 计算新的宽高
	if width > 0 && height > 0 {
		// 如果同时指定了宽度和高度，保持比例并调整
		if float64(image.Width())/float64(image.Height()) > float64(width)/float64(height) {
			height = int(float64(image.Height()) * float64(width) / float64(image.Width()))
		} else {
			width = int(float64(image.Width()) * float64(height) / float64(image.Height()))
		}
	} else if width > 0 {
		// 如果只指定了宽度，根据宽度计算高度
		height = int(float64(image.Height()) * float64(width) / float64(image.Width()))
	} else if height > 0 {
		// 如果只指定了高度，根据高度计算宽度
		width = int(float64(image.Width()) * float64(height) / float64(image.Height()))
	} else {
		return ErrMissingDimensions
	}

	// 统一调整图像大小
	scale := float64(width) / float64(image.Width())
	if float64(height) < float64(image.Height())*scale {
		scale = float64(height) / float64(image.Height())
	}

	if err = image.Resize(scale, vips.KernelLinear); err != nil {
		return err
	}

	ep := vips.NewDefaultJPEGExportParams()
	imageBytes, _, err := image.Export(ep)
	if err != nil {
		return err
	}

	return os.WriteFile(target, imageBytes, 0644)
}
