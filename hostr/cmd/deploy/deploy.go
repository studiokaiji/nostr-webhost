package deploy

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/keystore"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/relays"
	"golang.org/x/net/html"
)

func isExternalURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
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
		fmt.Println("❌ Failed to get all relays:", err)
		return "", "", "", err
	}

	// basePath以下のText Fileのパスをすべて羅列する
	err = generateEventsAndAddQueueAllValidStaticTextFiles(
		priKey,
		pubKey,
		htmlIdentifier,
		basePath,
		replaceable,
	)
	if err != nil {
		fmt.Println("❌ Failed to convert text files:", err)
		return "", "", "", err
	}

	// basePath以下のMedia Fileのパスを全て羅列しアップロード
	err = uploadAllValidStaticMediaFiles(priKey, pubKey, basePath)
	if err != nil {
		fmt.Println("❌ Failed to upload media:", err)
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

	addNostrEventQueue(event, filePath)

	eventId, encoded := publishEventsFromQueue(replaceable)

	return eventId, encoded, htmlIdentifier, err
}

func convertLinks(
	priKey, pubKey, basePath string,
	replaceable bool,
	indexHtmlIdentifier string,
	n *html.Node,
) {
	if n.Type == html.ElementNode {
		if n.Data == "link" || n.Data == "script" {
			// <link> と <script> タグを対象としてNostr Eventを作成
			for i, a := range n.Attr {
				// href,srcのうち、外部URLでないものかつ. html, .css, .js のみ変換する
				if (a.Key == "href" || a.Key == "src") && !isExternalURL(a.Val) && isValidBasicFileType(a.Val) {
					filePath := filepath.Join(basePath, a.Val)

					// kindを取得
					kind, err := pathToKind(filePath, replaceable)
					if err != nil {
						break
					}

					// contentを取得
					bytesContent, err := os.ReadFile(filePath)
					if err != nil {
						fmt.Println("❌ Failed to read", filePath, ":", err)
						continue
					}

					content := string(bytesContent)

					// Tagsを追加
					tags := nostr.Tags{}
					// 置き換え可能なイベントの場合
					if replaceable {
						fileIdentifier := getReplaceableIdentifier(indexHtmlIdentifier, a.Val)
						tags = tags.AppendUnique(nostr.Tag{"d", fileIdentifier})
						// 元のパスをfileIdentifierに置き換える
						n.Attr[i].Val = fileIdentifier
					}

					// jsファイルを解析する
					if strings.HasSuffix(a.Val, ".js") {
						// アップロード済みファイルの元パスとURLを取得
						for path, url := range uploadedMediaFilePathToURL {
							// JS内に該当ファイルがあったら置換
							content = strings.ReplaceAll(content, path, url)
						}
					}

					event, err := getEvent(priKey, pubKey, content, kind, tags)
					if err != nil {
						fmt.Println("❌ Failed to get event for", filePath, ":", err)
						break
					}

					addNostrEventQueue(event, filePath)

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
	}

	// 子ノードに対して再帰的に処理
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertLinks(priKey, pubKey, basePath, replaceable, indexHtmlIdentifier, c)
	}
}
