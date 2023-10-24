# Nostr Webhost (hostr)

Example webpage: https://h.hostr.cc/p/2d417bce8c10883803bc427703e3c4c024465c88e7063ed68f9dfeecf56911ac/d/hostr-lp

## 🌐 Overview

Nostr Webhost (hostr) is a command-line tool designed for hosting Single Page Applications (SPAs) using the Nostr protocol and its distributed network of relay servers. This tool provides a seamless way to deploy and access your SPA on the Nostr network.

## ⚠️ Caution
Domain-based authorization mechanisms such as NIP-7 should not currently be used. This is because the event is identified based on the path, so it will authorize other events as well.

### 📦 Installation

To get started with hostr, follow these steps:

1. `go install github.com/studiokaiji/nostr-webhost/hostr@latest`

### ⌨️ Commands

```bash
COMMANDS:
   deploy        🌐 Deploy nostr website
   add-relay     📌 Add nostr relay
   remove-relay  🗑 Remove nostr relay
   list-relay    📝 List added nostr relays
   set-private   🔐 Set private key
   show-public   📛 Show public key
   generate-key  🗝 Generate key
   start         🕺 Wake up web server
   help, h       Shows a list of commands or help for one command
```

### 🚀 Getting Started

1. Install Nostr Webhost as mentioned above.
2. Set or generate private key
If you set private key: `hostr set-private "nsec or hex private key"`
Or if you want to generate private key: `hostr generate-key`
3. Add relay
`hostr add-relay wss://r.hostr.cc`
4. Deploy
`hostr deploy --path /BUILT/SPA/DIR/PATH --identifier=test`
   - The `--identifier` option is the identifier (d-tag) for Replaceable Events based on NIP-33. When you update this site, please specify the same identifier. If you want to create a non-replaceable site, you can achieve that by specifying `--replaceable=false`.
   - The event id of index.html will be output after deploy. Please make a copy of it.
5. Start test web server
`hostr start`
6. Access the `http://localhost:3000/e/{event-id-of-index.html}`

For detailed information on how to use each command, you can use the `help` command followed by the specific command name.

## 👍 Feedback and Contributions

If you encounter any issues or have suggestions for improvement, feel free to contribute to the project on GitHub [link to GitHub repository].

## 📃 License

This project is licensed under the MIT. See the LICENSE file for more details.
