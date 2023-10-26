package deploy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/keystore"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/relays"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/tools"
	"golang.org/x/net/html"
)

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

func addNostrEventQueue(event *nostr.Event) {
	nostrEventsQueue = append(nostrEventsQueue, event)
}

var allRelays []string

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
	uploadedFilesCount := 0

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		tools.DisplayProgressBar(&uploadedFilesCount, &allEventsCount)
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
			mutex.Lock()         // ロックして排他制御
			uploadedFilesCount++ // カウントアップ
			mutex.Unlock()       // ロック解除
			wg.Done()            // ゴルーチンの終了を通知
		}(ev)
	}

	wg.Wait()

	if uploadedFilesCount < allEventsCount {
		fmt.Println("Failed to deploy", allEventsCount-uploadedFilesCount, "files.")
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

func isExternalURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidFileType(str string) bool {
	return strings.HasSuffix(str, ".html") || strings.HasSuffix(str, ".css") || strings.HasSuffix(str, ".js")
}

func Deploy(basePath string, replaceable bool, htmlIdentifier string) (string, string, string, error) {
	// 引数からデプロイしたいサイトのパスを受け取る。
	filePath := filepath.Join(basePath, "index.html")

	// パスのディレクトリ内のファイルからindex.htmlファイルを取得
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Failed to read index.html:", err)
		return "", "", "", err
	}

	// HTMLの解析
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		fmt.Println("❌ Failed to parse index.html:", err)
		return "", "", "", err
	}

	// Eventの取得に必要になるキーペアを取得
	priKey, err := keystore.GetSecret()
	if err != nil {
		fmt.Println("❌ Failed to get private key:", err)
		return "", "", "", err
	}
	pubKey, err := nostr.GetPublicKey(priKey)
	if err != nil {
		fmt.Println("❌ Failed to get public key:", err)
		return "", "", "", err
	}

	// htmlIdentifierの存在チェック
	if replaceable && len(htmlIdentifier) < 1 {
		// htmlIdentifierが指定されていない場合はユーザー入力を受け取る
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("⌨️ Please type identifier: ")

		htmlIdentifier, _ = reader.ReadString('\n')
		// 改行タグを削除
		htmlIdentifier = strings.TrimSpace(htmlIdentifier)

		fmt.Printf("Identifier: %s\n", htmlIdentifier)
	}

	// リレーを取得
	allRelays, err = relays.GetAllRelays()
	if err != nil {
		return "", "", "", err
	}

	// リンクの解析と変換
	convertLinks(priKey, pubKey, basePath, replaceable, htmlIdentifier, doc)

	// 更新されたHTML
	var buf bytes.Buffer
	html.Render(&buf, doc)

	strHtml := buf.String()

	// index.htmlのkindを設定
	indexHtmlKind := consts.KindWebhostHTML
	if replaceable {
		indexHtmlKind = consts.KindWebhostReplaceableHTML
	}

	// Tagsを追加
	tags := nostr.Tags{}
	if replaceable {
		tags = tags.AppendUnique(nostr.Tag{"d", htmlIdentifier})
	}

	// Eventを生成しキューに追加
	event, err := getEvent(priKey, pubKey, strHtml, indexHtmlKind, tags)
	if err != nil {
		fmt.Println("❌ Failed to get public key:", err)
		return "", "", "", err
	}
	addNostrEventQueue(event)
	fmt.Println("Added", filePath, "event to publish queue")

	eventId, encoded := publishEventsFromQueue(replaceable)

	return eventId, encoded, htmlIdentifier, err
}

func convertLinks(priKey, pubKey, basePath string, replaceable bool, indexHtmlIdentifier string, n *html.Node) {
	// <link> と <script> タグを対象とする
	if n.Type == html.ElementNode && (n.Data == "link" || n.Data == "script") {
		for i, a := range n.Attr {
			// href,srcのうち、外部URLでないものかつ. html, .css, .js のみ変換する
			if (a.Key == "href" || a.Key == "src") && !isExternalURL(a.Val) && isValidFileType(a.Val) {
				filePath := filepath.Join(basePath, a.Val)

				// kindを取得
				kind, err := pathToKind(filePath, replaceable)
				if err != nil {
					continue
				}

				// contentを取得
				bytesContent, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Println("❌ Failed to read", filePath, ":", err)
					continue
				}

				// Tagsを追加
				tags := nostr.Tags{}
				// 置き換え可能なイベントの場合
				if replaceable {
					fileIdentifier := getReplaceableIdentifier(indexHtmlIdentifier, a.Val)
					tags = tags.AppendUnique(nostr.Tag{"d", fileIdentifier})
					// 元のパスをfileIdentifierに置き換える
					n.Attr[i].Val = fileIdentifier
				}

				// Eventを生成し、キューに追加
				event, err := getEvent(priKey, pubKey, string(bytesContent), kind, tags)
				if err != nil {
					fmt.Println("❌ Failed to get event for", filePath, ":", err)
					break
				}

				addNostrEventQueue(event)
				fmt.Println("Added", filePath, "event to publish queue")

				// 置き換え可能なイベントでない場合
				if !replaceable {
					// neventを指定
					nevent, err := nip19.EncodeEvent(event.ID, allRelays, pubKey)
					if err != nil {
						fmt.Println("❌ Failed to encode event", filePath, ":", err)
						break
					}
					n.Attr[i].Val = nevent
				}
			}
		}
	}

	// 子ノードに対して再帰的に処理
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertLinks(priKey, pubKey, basePath, replaceable, indexHtmlIdentifier, c)
	}
}

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
