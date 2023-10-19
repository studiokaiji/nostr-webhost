package tools

import (
	"fmt"

	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
)

func GetContentType(kind int) (string, error) {
	switch kind {
	case consts.KindWebhostHTML | consts.KindWebhostReplaceableHTML:
		return "text/html; charset=utf-8", nil
	case consts.KindWebhostCSS | consts.KindWebhostReplaceableCSS:
		return "text/css; charset=utf-8", nil
	case consts.KindWebhostJS | consts.KindWebhostReplaceableJS:
		return "text/javascript; charset=utf-8", nil
	default:
		return "", fmt.Errorf("Invalid Kind")
	}
}
