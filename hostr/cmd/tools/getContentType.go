package tools

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
)

func GetContentType(event nostr.Event) (string, error) {
	kind := event.Kind

	if kind == "" {
		
	}

	if kind == consts.KindWebhostHTML || kind == consts.KindWebhostReplaceableHTML {
		return "text/html; charset=utf-8", nil
	} else if kind == consts.KindWebhostCSS || kind == consts.KindWebhostReplaceableCSS {
		return "text/css; charset=utf-8", nil
	} else if kind == consts.KindWebhostJS || kind == consts.KindWebhostReplaceableJS {
		return "text/javascript; charset=utf-8", nil
	} else {
		return "", fmt.Errorf("Invalid Kind")
	}
}
