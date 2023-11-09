package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/deploy"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/keystore"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/relays"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/server"
	"github.com/urfave/cli/v2"
)

//go:embed cute-ostrich.txt
var cuteOstrich string

func main() {
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
					&cli.BoolFlag{
						Name:    "replaceable",
						Aliases: []string{"r"},
						Usage:   "Specify 'true' explicitly when using NIP-33",
						Value:   true,
					},
					&cli.StringFlag{
						Name:    "identifier",
						Aliases: []string{"d"},
						Usage:   "index.html identifier (valid only if replaceable option is true)",
					},
				},
				Action: func(ctx *cli.Context) error {
					fmt.Println("🌐 Deploying...")

					path := ctx.String("path")
					replaceable := ctx.Bool("replaceable")
					dTag := ctx.String("identifier")

					_, encoded, dTag, err := deploy.Deploy(path, replaceable, dTag)
					if err == nil {
						fmt.Println("🌐 Deploy Complete!")

						pubkey, err := keystore.GetPublic()
						if err != nil {
							os.Exit(1)
						}
						npub, err := nip19.EncodePublicKey(pubkey)
						if err != nil {
							os.Exit(1)
						}

						defaultModeUrl := "https://h.hostr.cc"
						secureModeUrl := fmt.Sprintf("https://%s.hostr.cc", npub)

						if replaceable {
							defaultModeUrl = fmt.Sprintf("%s/p/%s/d/%s", defaultModeUrl, npub, dTag)
							secureModeUrl = fmt.Sprintf("%s/d/%s", secureModeUrl, dTag)
						} else {
							defaultModeUrl = fmt.Sprintf("%s/e/%s", defaultModeUrl, encoded)
							secureModeUrl = fmt.Sprintf("%s/e/%s", secureModeUrl, encoded)
						}

						fmt.Println("\n\033[1m========= 🚵 Access To =========\033")

						fmt.Printf("\n\033[1m🗃️  Default Mode:\033[0m\n\x1b[36m%s\x1b[0m\n", defaultModeUrl)
						fmt.Printf("\033[1m\n🔑 Secure Mode:\033[0m\n\x1b[36m%s\x1b[0m\n", secureModeUrl)

						fmt.Printf("\n\x1b[90mh.hostr.cc is just one endpoint, so depending on the relay configuration, it may not be accessible.\x1b[0m\n")

						fmt.Printf("\n\033[1m=================================\033\n")
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
						fmt.Print("🗝  Generated key\n🗝  You can check the public key with 'hostr show-public'\n")
					}
					return err
				},
			},
			{
				Name:  "start",
				Usage: "🕺 Wake up web server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   "3000",
						Usage:   "Web server port",
					},
					&cli.StringFlag{
						Name:    "mode",
						Aliases: []string{"m"},
						Value:   "normal",
						Usage:   "🧪 Experimental: Enabled subdomain-based access in replaceable events.",
						Action: func(ctx *cli.Context, v string) error {
							if v != "normal" && v != "hybrid" && v != "secure" {
								return fmt.Errorf("Invalid mode flag. Must be 'normal', 'hybrid', or 'secure'.")
							}
							return nil
						},
					},
				},
				Action: func(ctx *cli.Context) error {
					port := ctx.String("port")
					mode := ctx.String("mode")
					server.Start(port, mode)
					return nil
				},
			},
		},
	}

	if len(os.Args) < 2 || os.Args[1] == "help" || os.Args[1] == "h" {
		// Display ostrich
		fmt.Println(cuteOstrich)
	}

	// Start app
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
