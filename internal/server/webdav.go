package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

const (
	MethodHead      = "HEAD"
	MethodGet       = "GET"
	MethodPut       = "PUT"
	MethodPost      = "POST"
	MethodPatch     = "PATCH"
	MethodDelete    = "DELETE"
	MethodOptions   = "OPTIONS"
	MethodMkcol     = "MKCOL"
	MethodCopy      = "COPY"
	MethodMove      = "MOVE"
	MethodLock      = "LOCK"
	MethodUnlock    = "UNLOCK"
	MethodPropfind  = "PROPFIND"
	MethodProppatch = "PROPPATCH"
)

var WebDAVHandler = func(c *gin.Context, router *gin.RouterGroup, conf *config.Config, isOnlyRead bool) {
	var webDAVDir string
	username, _ := c.Get("username")

	if isOnlyRead {
		webDAVDir = filepath.Join(conf.StoragePath, "originals", username.(string))
	} else {
		webDAVDir = filepath.Join(conf.StoragePath, "user", username.(string))
	}

	if _, err := os.Stat(webDAVDir); os.IsNotExist(err) {
		err := os.MkdirAll(webDAVDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user directory"})
			return
		}
	}

	srv := &webdav.Handler{
		Prefix:     fmt.Sprintf("%s/%s", router.BasePath(), username),
		FileSystem: webdav.Dir(webDAVDir),
		LockSystem: webdav.NewMemLS(),
	}

	srv.ServeHTTP(c.Writer, c.Request)
}

func WebDAVOriginals(conf *config.Config, router *gin.RouterGroup) {
	handlerFunc := func(c *gin.Context) {
		WebDAVHandler(c, router, conf, true)
	}

	handleRead := func(h func(*gin.Context)) {
		router.Handle(MethodHead, "/*path", h)
		router.Handle(MethodGet, "/*path", h)
		router.Handle(MethodOptions, "/*path", h)
		router.Handle(MethodLock, "/*path", h)
		router.Handle(MethodUnlock, "/*path", h)
		router.Handle(MethodPropfind, "/*path", h)
	}

	handleRead(handlerFunc)
}

func WebDAVUser(conf *config.Config, router *gin.RouterGroup) {
	handlerFunc := func(c *gin.Context) {
		WebDAVHandler(c, router, conf, false)
	}

	handleRead := func(h func(*gin.Context)) {
		router.Handle(MethodHead, "/*path", h)
		router.Handle(MethodGet, "/*path", h)
		router.Handle(MethodOptions, "/*path", h)
		router.Handle(MethodLock, "/*path", h)
		router.Handle(MethodUnlock, "/*path", h)
		router.Handle(MethodPropfind, "/*path", h)
	}

	handleWrite := func(h func(*gin.Context)) {
		router.Handle(MethodPut, "/*path", h)
		router.Handle(MethodPost, "/*path", h)
		router.Handle(MethodPatch, "/*path", h)
		router.Handle(MethodDelete, "/*path", h)
		router.Handle(MethodMkcol, "/*path", h)
		router.Handle(MethodCopy, "/*path", h)
		router.Handle(MethodMove, "/*path", h)
		router.Handle(MethodProppatch, "/*path", h)
	}

	handleRead(handlerFunc)

	handleWrite(handlerFunc)
}

func WebDAVAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		username, password, err := parseUser(auth)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !checkAuth(username, password) {
			logger.Infof("check auth failed, username: %s, password: %s", username, password)
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Set("username", username)
		c.Next()
	}
}

func parseUser(auth string) (string, string, error) {
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", errors.New("parse auth failed")
	}

	payload, _ := base64.StdEncoding.DecodeString(parts[1])
	pair := strings.SplitN(string(payload), ":", 2)

	if len(pair) != 2 {
		return "", "", errors.New("parse auth failed")
	}

	return pair[0], pair[1], nil
}

func checkAuth(username, password string) bool {
	user := entity.FindUser(username)
	if user == nil {
		logger.Infof("user %s not found", username)
		return false
	}

	return user.InvalidPassword(password)
}
