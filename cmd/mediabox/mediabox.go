package main

import (
	"os"

	"github.com/skyline93/mediabox/internal/commands"
	"github.com/skyline93/mediabox/internal/log"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var logger *logrus.Logger

func init() {
	logger = log.NewLogger("app.log")
}

func main() {
	app := cli.NewApp()
	app.Name = "mediabox"
	app.Usage = "mediabox"
	app.Commands = commands.MediaBox

	if err := app.Run(os.Args); err != nil {
		logger.Error(err)
	}
}
