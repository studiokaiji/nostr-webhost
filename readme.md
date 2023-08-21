# Nostr Webhost

## Overview

Nostr Webhost is a command-line tool designed for hosting Single Page Applications (SPAs) using the Nostr protocol and its distributed network of relay servers. This tool provides a seamless way to deploy and access your SPA on the Nostr network.

### Installation

To get started with Nostr Webhost, follow these steps:

1. Install the tool using the following command:

2. Ensure you have Node.js and npm installed on your machine.

### Commands

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

### Getting Started

1. Install Nostr Webhost as mentioned above.
2. Set or generate private key
If you set private key: `nostrh set-private "nsec or hex private key"`
Or if you want to generate private key: `nostrh generate-key`
3. Add relay
`nostrh add-relay wss://nostrwebhost.studiokaiji.com`
4. Deploy
`nostrh deploy /BUILT/SPA/DIR/PATH`
The event id of index.html will be output after deploy. Please make a copy of it.
5. Start test web server
`nostrh start`
6. Access the `http://localhost:3000/e/{event-id-of-index.html}`

For detailed information on how to use each command, you can use the `help` command followed by the specific command name.

## Feedback and Contributions

If you encounter any issues or have suggestions for improvement, feel free to contribute to the project on GitHub [link to GitHub repository].

## License

This project is licensed under the MIT. See the LICENSE file for more details.
