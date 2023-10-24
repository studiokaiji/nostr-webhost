package deploy

import (
	"fmt"
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
