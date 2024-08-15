package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/studio-b12/gowebdav"
	"github.com/urfave/cli"
)

func uploadFile(filePath string, client *gowebdav.Client, targetPath string) error {
	remotePath := filePath[len(targetPath):]

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	err = client.Write(remotePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func WatchDirectory(client *gowebdav.Client, targetPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					err := uploadFile(event.Name, client, targetPath)
					if err != nil {
						log.Printf("Error uploading file: %v", err)
					} else {
						log.Printf("Uploaded: %s", event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
			}
		}
	}()

	err = watcher.Add(targetPath)
	if err != nil {
		log.Fatalf("failed to add directory: %v", err)
	}

	<-done
}

func startAction(c *cli.Context) error {
	log.Println("Starting WebDAV Client...")

	client := gowebdav.NewClient(
		c.String("server"),
		c.String("user"),
		c.String("password"),
	)

	err := client.Connect()
	if err != nil {
		log.Fatalf("failed to connect to WebDAV server: %v", err)
	}

	WatchDirectory(client, c.String("path"))

	return nil
}

var startFlags = []cli.Flag{
	cli.StringFlag{
		Name:  "server, S",
		Usage: "server url",
		Value: "http://127.0.0.1:8000/webdav/user/admin",
	},
	cli.StringFlag{
		Name:  "user, u",
		Usage: "user name",
	},
	cli.StringFlag{
		Name:  "password, p",
		Usage: "password",
	},
	cli.StringFlag{
		Name:  "path",
		Usage: "watch dir",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "agent"
	app.Usage = "agent"
	app.Commands = []cli.Command{
		cli.Command{
			Name:    "start",
			Aliases: []string{"up"},
			Usage:   "Starts the Web server",
			Flags:   startFlags,
			Action:  startAction,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("err: %s", err)
		os.Exit(1)
	}
}
