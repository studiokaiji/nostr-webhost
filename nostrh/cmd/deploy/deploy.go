package deploy

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/nostrh/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/nostrh/cmd/keystore"
	"github.com/studiokaiji/nostr-webhost/nostrh/cmd/relays"
	"github.com/studiokaiji/nostr-webhost/nostrh/cmd/tools"
	"golang.org/x/net/html"
)

func pathToKind(path string) (int, error) {
	splittedPath := strings.Split(path, ".")
	ex := splittedPath[len(splittedPath)-1]
	switch ex {
	case "html":
		return consts.KindWebhostHTML, nil
	case "css":
		return consts.KindWebhostCSS, nil
	case "js":
		return consts.KindWebhostJS, nil
	default:
		return 0, nil
	}
}

var nostrEventsQueue []*nostr.Event

func addNostrEventQueue(event *nostr.Event) {
	nostrEventsQueue = append(nostrEventsQueue, event)
}

func publishEventsFromQueue() (string, error) {
	ctx := context.Background()

	fmt.Println("Publishing...")

	// リレーを取得
	allRelays, err := relays.GetAllRelays()
	if err != nil {
		return "", err
	}

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

	indexEventId := nostrEventsQueue[len(nostrEventsQueue)-1].ID

	return indexEventId, err
}

func isExternalURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return false
}

func isValidFileType(str string) bool {
	return strings.HasSuffix(str, ".html") || strings.HasSuffix(str, ".css") || strings.HasSuffix(str, ".js")
}

func Deploy(basePath string) (string, error) {
	// 引数からデプロイしたいサイトのパスを受け取る。
	filePath := filepath.Join(basePath, "index.html")

	// パスのディレクトリ内のファイルからindex.htmlファイルを取得
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Failed to read index.html:", err)
		return "", err
	}

	// HTMLの解析
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		fmt.Println("❌ Failed to parse index.html:", err)
		return "", nil
	}

	// Eventの取得に必要になるキーペアを取得
	priKey, err := keystore.GetSecret()
	if err != nil {
		fmt.Println("❌ Failed to get private key:", err)
		return "", err
	}
	pubKey, err := nostr.GetPublicKey(priKey)
	if err != nil {
		fmt.Println("❌ Failed to get public key:", err)
		return "", err
	}

	// index.htmlファイル内に記述されている他Assetのパスから実際のデータを取得。
	// 取得してきたデータをeventにしてIDを取得。
	// リンクの解析と変換
	convertLinks(priKey, pubKey, basePath, doc)

	// 更新されたHTML
	var buf bytes.Buffer
	html.Render(&buf, doc)

	strHtml := buf.String()

	// Eventを生成しキューに追加
	event, err := getEvent(priKey, pubKey, strHtml, consts.KindWebhostHTML)
	if err != nil {
		fmt.Println("❌ Failed to get public key:", err)
		return "", err
	}
	addNostrEventQueue(event)
	fmt.Println("Added", filePath, "event to publish queue")

	return publishEventsFromQueue()
}

func convertLinks(priKey, pubKey, basePath string, n *html.Node) {
	// <link> と <script> タグを対象とする
	if n.Type == html.ElementNode && (n.Data == "link" || n.Data == "script") {
		for i, a := range n.Attr {
			// href,srcのうち、外部URLでないものかつ. html, .css, .js のみ変換する
			if (a.Key == "href" || a.Key == "src") && !isExternalURL(a.Val) && isValidFileType(a.Val) {
				filePath := filepath.Join(basePath, a.Val)
				// kindを取得
				kind, err := pathToKind(filePath)
				if err != nil {
					continue
				}

				// contentを取得
				bytesContent, err := os.ReadFile(filePath)
				if err != nil {
					fmt.Println("❌ Failed to read", filePath, ":", err)
					continue
				}

				// Eventを生成し、キューに追加
				event, err := getEvent(priKey, pubKey, string(bytesContent), kind)
				if err != nil {
					fmt.Println("❌ Failed to get event for", filePath, ":", err)
					break
				}

				addNostrEventQueue(event)
				fmt.Println("Added", filePath, "event to publish queue")

				// 元のパスをEvent[.]IDに変更
				n.Attr[i].Val = event.ID
			}
		}
	}

	// 子ノードに対して再帰的に処理
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertLinks(priKey, pubKey, basePath, c)
	}
}

func getEvent(priKey, pubKey, content string, kind int) (*nostr.Event, error) {
	ev := nostr.Event{
		PubKey:    pubKey,
		CreatedAt: nostr.Now(),
		Kind:      kind,
		Content:   content,
	}

	err := ev.Sign(priKey)
	if err != nil {
		return nil, err
	}

	return &ev, err
}
