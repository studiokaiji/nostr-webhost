package deploy

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/tools"
)

var availableContentTypes = []string{
	"image/png",
	"image/jpg",
	"image/jpeg",
	"image/gif",
	"image/webp",
	"video/mp4",
	"video/quicktime",
	"video/mpeg",
	"video/webm",
	"audio/mpeg",
	"audio/mpg",
	"audio/mpeg3",
	"audio/mp3",
}

var availableContentSuffixes = []string{
	".png",
	".jpg",
	".jpeg",
	".gif",
	".webp",
	".mp4",
	".quicktime",
	".mpeg",
	".webm",
	".mpeg",
	".mpg",
	".mpeg3",
	".mp3",
}

var availableMediaHtmlTags = []string{
	"img",
	"audio",
	"video",
	"source",
	"object",
	"embed",
}

func isValidMediaFileType(path string) bool {
	for _, suffix := range availableContentSuffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

const uploadEndpoint = "https://nostrcheck.me/api/v1/media"

type MediaResult struct {
	result      bool
	description string
	status      string
	id          int
	pubkey      string
	url         string
	hash        string
	magnet      string
	tags        []string
}

var mediaUploadRequestQueue []func() (*MediaResult, error)

func addNostrEventQueue(event *nostr.Event) {
	nostrEventsQueue = append(nostrEventsQueue, event)
}

func addMediaUploadRequestFuncQueue(reqFunc func() (*MediaResult, error)) {
	mediaUploadRequestQueue = append(mediaUploadRequestQueue, reqFunc)
}

var allRelays []string

func uploadMediaFilesFromQueue() {
	// Publishの進捗状況を表示
	allEventsCount := len(mediaUploadRequestQueue)
	uploadedFilesCount := 0

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		tools.DisplayProgressBar(&uploadedFilesCount, &allEventsCount)
		wg.Done()
	}()

	var mutex sync.Mutex

	// アップロードを並列処理
	for _, reqFunc := range mediaUploadRequestQueue {
		wg.Add(1)
		go func(reqFun func() (*MediaResult, error)) {
			_, err := reqFun()
			if err != nil {
				fmt.Println(err)
				return
			}
			mutex.Lock()         // ロックして排他制御
			uploadedFilesCount++ // カウントアップ
			mutex.Unlock()       // ロック解除
			wg.Done()            // ゴルーチンの終了を通知
		}(reqFunc)
	}

	wg.Wait()
}

func filePathToUploadMediaRequest(filePath, priKey, pubKey string) (*http.Request, error) {
	// ファイルを開く
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s: %w", filePath, err)
	}
	defer file.Close()

	// リクエストボディのバッファを初期化
	var requestBody bytes.Buffer
	// multipart writerを作成
	writer := multipart.NewWriter(&requestBody)

	// uploadtypeフィールドを設定
	err = writer.WriteField("uploadtype", "media")
	if err != nil {
		return nil, fmt.Errorf("Error writing field: %w", err)
	}

	// mediafileフィールドを作成
	part, err := writer.CreateFormFile("mediafile", filePath)
	if err != nil {
		return nil, fmt.Errorf("Error creating form file: %w", err)
	}

	// ファイルの内容をpartにコピー
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("Error copying file: %w", err)
	}

	// writerを閉じてリクエストボディを完成させる
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("Error closing writer: %w", err)
	}

	// タグを初期化
	tags := nostr.Tags{}
	// タグを追加
	tags.AppendUnique(nostr.Tag{"u", uploadEndpoint})
	tags.AppendUnique(nostr.Tag{"method", "POST"})
	tags.AppendUnique(nostr.Tag{"payload", ""})

	// イベントを生成
	ev, err := getEvent(priKey, pubKey, "", 27533, tags)
	if err != nil {
		return nil, fmt.Errorf("Error get event: %d", err)
	}

	// イベントをJSONにマーシャル
	evJson, err := ev.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Error marshaling event: %d", err)
	}

	// HTTPリクエストを作成
	request, err := http.NewRequest("POST", uploadEndpoint, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %d", err)
	}

	// ヘッダーを設定
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString(evJson))
	request.Header.Set("Accept", "application/json")

	return request, nil
}
