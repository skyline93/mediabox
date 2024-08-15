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
		Name:  "host, H",
		Usage: "server host",
		Value: "127.0.0.1",
	},
	cli.IntFlag{
		Name:  "port, p",
		Usage: "server port",
		Value: 8000,
	},
	cli.StringFlag{
		Name:  "tls-cert",
		Usage: "tls cert file path",
		Value: "server.crt",
	},
	cli.StringFlag{
		Name:  "tls-key",
		Usage: "tls key file path",
		Value: "server.key",
	},
}

func startAction(c *cli.Context) error {
	conf := config.Config{
		HttpHost: c.String("host"),
		HttpPort: c.Int("port"),
		TLSCert:  c.String("tls-cert"),
		TLSKey:   c.String("tls-key"),
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
