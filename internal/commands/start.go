package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/skyline93/mediabox/internal/config"
	"github.com/skyline93/mediabox/internal/entity"
	"github.com/skyline93/mediabox/internal/server"

	"github.com/urfave/cli"
)

var StartCommand = cli.Command{
	Name:    "start",
	Aliases: []string{"up"},
	Usage:   "Starts the Web server",
	Flags:   startFlags,
	Action:  startAction,
}

var startFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "config",
		Usage: "options from config file",
	},
	cli.StringFlag{
		Name:   "host, H",
		Usage:  "server host",
		Value:  "127.0.0.1",
		EnvVar: "MEDIABOX_HOST",
	},
	cli.IntFlag{
		Name:   "port, p",
		Usage:  "server port",
		Value:  8000,
		EnvVar: "MEDIABOX_PORT",
	},
	cli.BoolFlag{
		Name:   "use-tls",
		Usage:  "is use tls",
		EnvVar: "MEDIABOX_USE_TLS",
	},
	cli.StringFlag{
		Name:   "tls-cert",
		Usage:  "tls cert file path",
		Value:  "server.crt",
		EnvVar: "MEDIABOX_TLS_CERT",
	},
	cli.StringFlag{
		Name:   "tls-key",
		Usage:  "tls key file path",
		Value:  "server.key",
		EnvVar: "MEDIABOX_TLS_KEY",
	},
	cli.StringFlag{
		Name:   "db-driver",
		Usage:  "db driver",
		EnvVar: "MEDIABOX_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "db-dsn",
		Usage:  "db dsn",
		EnvVar: "MEDIABOX_DATABASE_DSN",
	},
	cli.StringFlag{
		Name:   "storage-path",
		Usage:  "storage path",
		EnvVar: "MEDIABOX_STORAGE_PATH",
	},
}

func startAction(c *cli.Context) error {
	conf := config.Config{
		HttpHost:    c.String("host"),
		HttpPort:    c.Int("port"),
		UseTLS:      c.Bool("use-tls"),
		TLSCert:     c.String("tls-cert"),
		TLSKey:      c.String("tls-key"),
		DbDriver:    c.String("db-driver"),
		DbDsn:       c.String("db-dsn"),
		StoragePath: c.String("storage-path"),
	}

	configPath := c.String("config")

	if configPath != "" {
		configData, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(configData, &conf); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	entity.InitDb(&conf)
	server.Start(&conf)

	return nil
}
