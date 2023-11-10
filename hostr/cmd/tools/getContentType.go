package tools

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
)

// Content-Typeをイベントから取得する。NIP-95の場合は第二引数がtrueになる。
func GetContentType(event *nostr.Event) (string, bool, error) {
	kind := event.Kind

	if kind == consts.KindTextFile || kind == consts.KindReplaceableTextFile {
		contentTypeTag := event.Tags.GetFirst([]string{"type"})
		contentType := contentTypeTag.Value()

		if len(contentType) < 1 {
			return "", true, fmt.Errorf("Content-Type not specified")
		}

		return contentType, true, nil
	}

	if kind == consts.KindWebhostHTML || kind == consts.KindWebhostReplaceableHTML {
		return "text/html; charset=utf-8", false, nil
	} else if kind == consts.KindWebhostCSS || kind == consts.KindWebhostReplaceableCSS {
		return "text/css; charset=utf-8", false, nil
	} else if kind == consts.KindWebhostJS || kind == consts.KindWebhostReplaceableJS {
		return "text/javascript; charset=utf-8", false, nil
	} else {
		return "", false, fmt.Errorf("Invalid Kind")
	}
}
