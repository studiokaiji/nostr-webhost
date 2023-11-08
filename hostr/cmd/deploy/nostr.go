package deploy

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/tools"
)

var allRelays []string

func getEvent(priKey, pubKey, content string, kind int, tags nostr.Tags) (*nostr.Event, error) {
	ev := nostr.Event{
		PubKey:    pubKey,
		CreatedAt: nostr.Now(),
		Kind:      kind,
		Content:   content,
		Tags:      tags,
	}

	err := ev.Sign(priKey)
	if err != nil {
		return nil, err
	}

	return &ev, err
}

func isValidBasicFileType(str string) bool {
	return strings.HasSuffix(str, ".html") || strings.HasSuffix(str, ".css") || strings.HasSuffix(str, ".js")
}
func publishEventsFromQueue(replaceable bool) (string, string) {
	ctx := context.Background()

	fmt.Println("Publishing...")

	// 各リレーに接続
	var relays []*nostr.Relay

	for _, url := range allRelays {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			fmt.Println("❌ Failed to connect to:", url)
			continue
		}
		relays = append(relays, relay)
	}

	// Publishの進捗状況を表示
	allEventsCount := len(nostrEventsQueue)
	uploadedMediaFilePathToURLCount := 0

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		tools.DisplayProgressBar(&uploadedMediaFilePathToURLCount, &allEventsCount)
		wg.Done()
	}()

	var mutex sync.Mutex

	// リレーへpublish
	for _, ev := range nostrEventsQueue {
		wg.Add(1)
		go func(event *nostr.Event) {
			for _, relay := range relays {
				_, err := relay.Publish(ctx, *event)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			mutex.Lock()                      // ロックして排他制御
			uploadedMediaFilePathToURLCount++ // カウントアップ
			mutex.Unlock()                    // ロック解除
			wg.Done()                         // ゴルーチンの終了を通知
		}(ev)
	}

	wg.Wait()

	if uploadedMediaFilePathToURLCount < allEventsCount {
		fmt.Println("Failed to deploy", allEventsCount-uploadedMediaFilePathToURLCount, "files.")
	}

	indexEvent := nostrEventsQueue[len(nostrEventsQueue)-1]

	encoded := ""
	if !replaceable {
		if enc, err := nip19.EncodeEvent(indexEvent.ID, allRelays, indexEvent.PubKey); err == nil {
			encoded = enc
		} else {
			fmt.Println("❌ Failed to covert nevent:", err)
		}
	}

	return indexEvent.ID, encoded
}

func pathToKind(path string, replaceable bool) (int, error) {
	// パスを分割
	separatedPath := strings.Split(path, ".")
	// 拡張子を取得
	ex := separatedPath[len(separatedPath)-1]
	// replaceable(NIP-33)の場合はReplaceableなkindを返す
	switch ex {
	case "html":
		if replaceable {
			return consts.KindWebhostHTML, nil
		} else {
			return consts.KindWebhostReplaceableHTML, nil
		}
	case "css":
		if replaceable {
			return consts.KindWebhostReplaceableCSS, nil
		} else {
			return consts.KindWebhostCSS, nil
		}
	case "js":
		if replaceable {
			return consts.KindWebhostReplaceableJS, nil
		} else {
			return consts.KindWebhostJS, nil
		}
	default:
		return 0, fmt.Errorf("Invalid path")
	}
}

// Replaceableにする場合のidentifier(dタグ)を取得
func getReplaceableIdentifier(indexHtmlIdentifier, filePath string) string {
	return indexHtmlIdentifier + "/" + filePath[1:]
}

var nostrEventsQueue []*nostr.Event

func addNostrEventQueue(event *nostr.Event, filePath string) {
	nostrEventsQueue = append(nostrEventsQueue, event)
	fmt.Println("Added", filePath, "event to publish queue")
}
