package deploy

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
	Result      bool     `json:"result,omitempty"`
	Description string   `json:"description"`
	Status      string   `json:"status,omitempty"`
	Id          int      `json:"id,omitempty"`
	Pubkey      string   `json:"pubkey,omitempty"`
	Url         string   `json:"url,omitempty"`
	Hash        string   `json:"hash,omitempty"`
	Magnet      string   `json:"magnet,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// [元パス]:[URL]の形で記録する
var uploadedMediaFiles = map[string]string{}

func uploadMediaFiles(filePaths []string, requests []*http.Request) {
	fmt.Println("Uploading media files...")

	client := &http.Client{}

	var uploadedMediaFilesCount = 0
	var allMediaFilesCount = len(requests)

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		tools.DisplayProgressBar(&uploadedMediaFilesCount, &allMediaFilesCount)
		wg.Done()
	}()

	var mutex sync.Mutex

	// アップロードを並列処理
	for i, req := range requests {
		wg.Add(1)
		filePath := filePaths[i]

		fmt.Printf("\nAdded upload request %s", filePath)

		go func(filePath string, req *http.Request) {
			defer wg.Done()

			response, err := client.Do(req)
			// リクエストを送信
			if err != nil {
				fmt.Println("\n❌ Error sending request:", filePath, err)
				return
			}
			defer response.Body.Close()

			if !strings.HasPrefix(fmt.Sprint(response.StatusCode), "2") {
				fmt.Println("\n❌ Failed to upload:", response.StatusCode, filePath)
				return
			}

			var result *MediaResult
			// ResultのDecode
			err = json.NewDecoder(response.Body).Decode(&result)

			if err != nil {
				fmt.Println("\n❌ Error decoding response:", err)
				return
			}

			// アップロードに失敗した場合
			if !result.Result {
				fmt.Println("\n❌ Failed to upload file:", filePath, err)
				return
			}

			mutex.Lock()              // ロックして排他制御
			uploadedMediaFilesCount++ // カウントアップ
			uploadedMediaFiles[filePath] = result.Url
			mutex.Unlock() // ロック解除
		}(filePath, req)
	}

	wg.Wait()
}

func filePathToUploadMediaRequest(basePath, filePath, priKey, pubKey string) (*http.Request, error) {
	// ファイルを開く
	file, err := os.Open(filepath.Join(basePath, filePath))
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s: %w", filePath, err)
	}
	defer file.Close()

	// リクエストボディのバッファを初期化
	var requestBody bytes.Buffer
	// multipart writerを作成
	writer := multipart.NewWriter(&requestBody)

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

	// uploadtypeフィールドを設定
	err = writer.WriteField("uploadtype", "media")
	if err != nil {
		return nil, fmt.Errorf("Error writing field: %w", err)
	}

	// writerを閉じてリクエストボディを完成させる
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("Error closing writer: %w", err)
	}

	// タグを追加
	tags := nostr.Tags{
		nostr.Tag{"u", uploadEndpoint},
		nostr.Tag{"method", "POST"},
		nostr.Tag{"payload", ""},
	}

	// イベントを生成
	ev, err := getEvent(priKey, pubKey, "", 27235, tags)
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
	request.Header.Set("Authorization", "Nostr "+base64.StdEncoding.EncodeToString(evJson))
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", writer.FormDataContentType())

	return request, nil
}

// basePath以下のMedia Fileのパスを全て羅列する
func listAllValidStaticMediaFilePaths(basePath string) ([]string, error) {
	mediaFilePaths := []string{}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if !info.IsDir() {
			// 各サフィックスに対してマッチングを試みる
			for _, suffix := range availableContentSuffixes {
				// ファイル名とサフィックスがマッチした場合
				if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(suffix)) {
					// フルパスからbasePathまでの相対パスを計算
					relPath, err := filepath.Rel(basePath, path)
					if err != nil {
						fmt.Println("❌ Error calculating relative path:", err)
						continue
					}
					// マッチするファイルの相対パスをスライスに追加
					mediaFilePaths = append(mediaFilePaths, "/"+relPath)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return mediaFilePaths, nil
}

// basePath以下のMedia Fileのパスを全て羅列しアップロード
func uploadAllValidStaticMediaFiles(priKey, pubKey, basePath string) error {
	filesPaths, err := listAllValidStaticMediaFilePaths(basePath)
	if err != nil {
		return err
	}

	requests := []*http.Request{}

	for _, filePath := range filesPaths {
		request, err := filePathToUploadMediaRequest(basePath, filePath, priKey, pubKey)
		if err != nil {
			return err
		}
		requests = append(requests, request)
	}

	uploadMediaFiles(filesPaths, requests)

	return nil
}
