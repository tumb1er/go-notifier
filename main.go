package main

import (
	"github.com/urfave/cli"
	"go-notifier/notifier"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func watch(address string, icon string) error {
	n, err := notifier.NewNotifier(icon)
	if err != nil {
		return err
	}
	defer n.Close()

	err = n.AddNotifyIcon("notifier", "started", "listening for events")
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
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func(){
		<- c
		// suppress unrelated log.Print in notifier.eventLoop
		log.SetOutput(ioutil.Discard)
		if err := observer.Stop(); err != nil {
			log.SetOutput(os.Stdout)
			log.Fatalf("stop error: %e", err)
		}
	}()
	if err:= observer.Observe(address, handle); err != nil{
		return err
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "notifier"
	app.Usage = "listen for notifications"
	app.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "address",
			Value: "localhost:9998",
			Usage: "tcp address for socket server",
		},
		cli.StringFlag{
			Name: "icon",
			Value: "icon.ico",
			Usage: "notification icon path",
		},
	}
	app.Action = func(c *cli.Context) error {
		return watch(c.String("address"), c.String("icon"))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
