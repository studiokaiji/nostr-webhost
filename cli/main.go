package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/cmd/deploy"
	"github.com/studiokaiji/nostr-webhost/cmd/keystore"
	"github.com/studiokaiji/nostr-webhost/cmd/relays"
	"github.com/studiokaiji/nostr-webhost/cmd/server"
	"github.com/urfave/cli/v2"
)

func main() {
	var (
		port string
	)

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "deploy",
				Usage: "🌐 Deploy nostr website",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "path",
						Aliases: []string{"p"},
						Value:   "./",
						Usage:   "Site directory",
					},
				},
				Action: func(ctx *cli.Context) error {
					fmt.Println("🌐 Deploying...")
					path := ctx.String("path")
					indexEventId, err := deploy.Deploy(path)
					if err == nil {
						fmt.Println("🌐 Deploy Complete!")
						fmt.Println("🌐 index.html Event ID:", indexEventId)
					}
					return err
				},
			},
			{
				Name:  "add-relay",
				Usage: "📌 Add nostr relay",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args()
					relay := args.Get(args.Len() - 1)
					err := relays.AddRelay(relay)
					if err == nil {
						fmt.Println("📌 Added relay:", relay)
					}
					return err
				},
			},
			{
				Name:  "remove-relay",
				Usage: "🗑  Remove nostr relay",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args()
					relay := args.Get(args.Len() - 1)
					err := relays.RemoveRelay(relay)
					if err == nil {
						fmt.Println("🗑  Removed relay:", relay)
					}
					return err
				},
			},
			{
				Name:  "list-relays",
				Usage: "📝 List added nostr relays",
				Action: func(ctx *cli.Context) error {
					relays, err := relays.GetAllRelays()
					fmt.Println("===========================")
					for _, relay := range relays {
						fmt.Println(relay)
					}
					fmt.Println("===========================")
					return err
				},
			},
			{
				Name:  "set-private",
				Usage: "🔐 Set private key",
				Action: func(ctx *cli.Context) error {
					args := ctx.Args()
					key := args.Get(args.Len() - 1)
					err := keystore.SetSecret(key)
					if err == nil {
						fmt.Println("🔐 Secret is recorded")
					}
					return err
				},
			},
			{
				Name:  "show-public",
				Usage: "📛 Show public key",
				Action: func(ctx *cli.Context) error {
					_, _, err := keystore.ShowPublic()
					return err
				},
			},
			{
				Name:  "generate-key",
				Usage: "🗝  Generate key",
				Action: func(ctx *cli.Context) error {
					key := nostr.GeneratePrivateKey()
					err := keystore.SetSecret(key)
					if err == nil {
						fmt.Print("🗝  Generated key\n🗝  You can check the public key with 'noshost show-public'\n")
					}
					return err
				},
			},
			{
				Name:  "start",
				Usage: "🕺 Wake up web server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "port",
						Aliases:     []string{"p"},
						Value:       "3000",
						Usage:       "Web server port",
						Destination: &port,
					},
				},
				Action: func(ctx *cli.Context) error {
					server.Start(port)
					return nil
				},
			},
		},
	}

	if len(os.Args) < 2 || os.Args[1] == "help" || os.Args[1] == "h" {
		// Display ostrich
		file, err := os.Open("./cute-ostrich.txt")
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		buf := new(bytes.Buffer)
		for {
			n, err := buf.ReadFrom(file)
			if n == 0 {
				break
			}
			if err != nil {
				panic(err)
			}

			fmt.Println(buf.String())
		}

	}

	// Start app
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
