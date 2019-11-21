package main

import (
	"fmt"
	"github.com/tumb1er/go-notifier/notifier"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func watch(address string, icon string, name string) error {
	n, err := notifier.NewNotifier(icon)
	if err != nil {
		return err
	}
	defer n.Close()

	err = n.AddNotifyIcon(name,
		fmt.Sprintf("%s started.", name),
		fmt.Sprintf("%s is listening for events.", name))
	if err != nil {
		return err
	}

	handle := func(tip, title, info string) {
		if err := n.Update(tip, title, info); err != nil {
			log.Fatalf("notification error: %e", err)
		}
	}

	observer := new(notifier.SocketTransport)
	c := make(chan os.Signal, 1)
	running := true
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		running = false
		// suppress unrelated log.Print in notifier.eventLoop
		log.SetOutput(ioutil.Discard)
		if err := observer.Stop(); err != nil {
			log.SetOutput(os.Stdout)
			log.Fatalf("stop error: %e", err)
		}
	}()

	for {
		if err := observer.Observe(address, handle); err != nil {
			log.Printf("observe error: %v", err)
			time.Sleep(time.Second)
		}
		if !running {
			return nil
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "notifier"
	app.Usage = "listen for notifications"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "address",
			Value: "localhost:9998",
			Usage: "tcp address for socket server",
		},
		&cli.StringFlag{
			Name:  "icon",
			Value: "icon.ico",
			Usage: "notification icon path",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "tray icon name",
			Value: "Notifier",
		},
	}
	app.Action = func(c *cli.Context) error {
		return watch(c.String("address"), c.String("icon"), c.String("name"))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
