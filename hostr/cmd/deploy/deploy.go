package deploy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/keystore"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/relays"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
)

func isExternalURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func Deploy(basePath string, replaceable bool, htmlIdentifier string) (string, string, error) {
	// 引数からデプロイしたいサイトのパスを受け取る。
	filePath := filepath.Join(basePath, "index.html")

	// パスのディレクトリ内のファイルからindex.htmlファイルを取得
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("❌ Failed to read index.html:", err)
		return "", "", err
	}

	// HTMLの解析
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		fmt.Println("❌ Failed to parse index.html:", err)
		return "", "", err
	}

	// Eventの取得に必要になるキーペアを取得
	priKey, err := keystore.GetSecret()
	if err != nil {
		fmt.Println("❌ Failed to get private key:", err)
		return "", "", err
	}
	pubKey, err := nostr.GetPublicKey(priKey)
	if err != nil {
		fmt.Println("❌ Failed to get public key:", err)
		return "", "", err
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
		return "", "", err
	}

	// リンクの解析と変換
	convertLinks(priKey, pubKey, basePath, replaceable, htmlIdentifier, doc)

	if len(mediaUploadRequestQueue) > 0 {
		// メディアのアップロード
		fmt.Println("📷 Uploading media files")
		uploadMediaFilesFromQueue()
		fmt.Println("📷 Media upload finished.")
	}

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
		return "", "", err
	}
	addNostrEventQueue(event)
	fmt.Println("Added", filePath, "event to publish queue")

	eventId, encoded := publishEventsFromQueue(replaceable)

	return eventId, encoded, err
}

func convertLinks(priKey, pubKey, basePath string, replaceable bool, indexHtmlIdentifier string, n *html.Node) {
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
						continue
					}

					// contentを取得
					bytesContent, err := os.ReadFile(filePath)
					if err != nil {
						fmt.Println("❌ Failed to read", filePath, ":", err)
						continue
					}

					// jsファイルを解析する
					if strings.HasSuffix(basePath, ".js") {
						jsContent := string(bytesContent)
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
		} else if slices.Contains(availableMediaHtmlTags, n.Data) {
			// 内部mediaファイルを対象にUpload Requestを作成
			for i, a := range n.Attr {
				if (a.Key == "href" || a.Key == "src" || a.Key == "data") && !isExternalURL(a.Val) && isValidMediaFileType(a.Val) {
					filePath := filepath.Join(basePath, a.Val)

					// アップロードのためのHTTPリクエストを取得
					request, err := filePathToUploadMediaRequest(filePath, priKey, pubKey)
					if err != nil {
						fmt.Println("❌ Failed generate upload request: ", err)
					}

					// アップロード処理を代入
					uploadFunc := func(client *http.Client) (*MediaResult, error) {
						response, err := client.Do(request)
						// リクエストを送信
						if err != nil {
							return nil, fmt.Errorf("Error sending request: %w", err)
						}
						defer response.Body.Close()

						var result *MediaResult
						// ResultのDecode
						err = json.NewDecoder(response.Body).Decode(result)
						if err != nil {
							return nil, fmt.Errorf("Error decoding response: %w", err)
						}

						// アップロードに失敗した場合
						if !result.result {
							return nil, fmt.Errorf("Failed to upload file: %w", err)
						}

						// URLを割り当て
						n.Attr[i].Val = result.url

						return result, nil
					}

					// Queueにアップロード処理を追加
					addMediaUploadRequestFuncQueue(uploadFunc)
				}
			}
		}
	}

	// 子ノードに対して再帰的に処理
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		convertLinks(priKey, pubKey, basePath, replaceable, indexHtmlIdentifier, c)
	}
}

func convertLinksFromJS() {

}
