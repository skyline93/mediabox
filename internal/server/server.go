package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skyline93/mediabox/internal/api"
	"github.com/skyline93/mediabox/internal/auth"
	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/log"
	"github.com/skyline93/mediabox/internal/mediabox"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	BaseUri = "/api/v1"
)

var (
	logger *logrus.Logger
	APIv1  *gin.RouterGroup
)

func init() {
	logger = log.NewLogger("server.log")
}

func registerRoutes(router *gin.Engine, conf *config.Config) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	WebDAVOriginals(conf, router.Group("/webdav/originals", WebDAVAuth()))
	WebDAVUser(conf, router.Group("/webdav/user", WebDAVAuth()))

	api.Login(router)
	api.Register(router)
	api.DeleteUser(router)

	api.Ping(APIv1)
	api.UploadFile(APIv1, conf)
	api.UploadPhoto(APIv1, conf)
	api.ImportPhoto(APIv1, conf)
	api.ListPhotos(APIv1, conf)
	api.GetPhotoFile(APIv1, conf)

	api.CreateAlbum(APIv1, conf)
	api.ListAlbums(APIv1, conf)
}

func StartHttp(ctx context.Context, conf *config.Config) {
	router := gin.Default()
	router.Use(api.ErrorHandler(), cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                                 // 允许访问的来源域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},           // 允许的请求方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"}, // 允许的请求头
		AllowCredentials: true,                                                          // 允许携带凭证（如 Cookies）
	}))

	APIv1 = router.Group(BaseUri, auth.AuthMiddleware())

	registerRoutes(router, conf)

	tcpSocket := fmt.Sprintf("%s:%d", conf.HttpHost, conf.HttpPort)

	listener, err := net.Listen("tcp", tcpSocket)
	if err != nil {
		logger.Errorf("server: listener %s", err)
		return
	}

	logger.Infof("server listen at %s", tcpSocket)

	server := &http.Server{Addr: tcpSocket, Handler: router}

	go func() {
		var err error
		if conf.UseTLS {
			err = server.ServeTLS(listener, conf.TLSCert, conf.TLSKey)
		} else {
			err = server.Serve(listener)
		}

		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Infof("server: shutdown complete")
			} else {
				logger.Infof("server: %s", err)
			}
		}
	}()

	<-ctx.Done()
	logger.Info("server: shutting down")

	if err = server.Close(); err != nil {
		logger.Errorf("server: shutdown failed (%s)", err)
	}
}

func Start(conf *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("start server")
	go StartHttp(ctx, conf)

	mediabox.VipsInit()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)

	sig := <-quit

	logger.Info("shutting down...")
	cancel()

	mediabox.VipsShutdown()

	time.Sleep(2 * time.Second)

	if sig == syscall.SIGUSR1 {
		os.Exit(1)
	}
}
