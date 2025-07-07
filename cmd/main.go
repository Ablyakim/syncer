package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

const defaultAddr = "0.0.0.0:8880"

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "server",
				Aliases: []string{"s"},
				Usage:   "Websocket server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Aliases: []string{"a"},
						Value:   defaultAddr,
						Usage:   "Address to listen to",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "path",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					server(cmd.String("addr"), cmd.StringArg("path"))
					return nil
				},
			},
			{
				Name:    "client",
				Aliases: []string{"c"},
				Usage:   "complete a task on the list",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "addr",
						Aliases: []string{"a"},
						Value:   defaultAddr,
						Usage:   "Address to listen to",
					},
				},
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name: "path",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					client(cmd.String("addr"), cmd.StringArg("path"))
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func listenOSKillSignals(ctx context.Context) context.Context {
	var cancelFunc context.CancelFunc
	ctx, cancelFunc = context.WithCancel(ctx)
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		select {
		case <-ch:
			cancelFunc()
		case <-ctx.Done():
			signal.Reset()
			return
		}
	}()

	return ctx
}
