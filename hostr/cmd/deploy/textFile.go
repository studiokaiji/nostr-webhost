package deploy

import (
	"encoding/base64"
	"os"
	"path/filepath"

	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/tools"
)

// 有効なText Fileの拡張子
var availableTextFileSuffixes = []string{
	".txt",
	".csv",
	".pdf",
	".json",
	".yml",
	".yaml",
	".svg",
}

// [Text Fileの拡張子]: Content-Typeで記録する
var availableTextFileContentTypes = map[string]string{
	".txt":  "text/plain",
	".csv":  "text/csv",
	".pdf":  "application/pdf",
	".json": "application/json",
	".yml":  "application/x-yaml",
	".yaml": "application/x-yaml",
	".svg":  "image/svg+xml",
}

// [元パス]:[event]の形で記録する
var textFilePathToEvent = map[string]*nostr.Event{}

// basePath以下のText Fileのパスを全て羅列する
func listAllValidStaticTextFiles(basePath string) ([]string, error) {
	return tools.FindFilesWithBasePathBySuffixes(basePath, availableTextFileSuffixes)
}

// Text fileをBase pathから割り出して、eventを生成しキューに追加
func generateEventsAndAddQueueAllValidStaticTextFiles(priKey, pubKey, indexHtmlIdentifier, basePath string, replaceable bool) error {
	filePaths, err := listAllValidStaticTextFiles(basePath)
	if err != nil {
		return err
	}

	for _, filePath := range filePaths {
		// ファイルを開く
		bytesContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		// ファイル内容をbase64エンコード
		content := base64.StdEncoding.EncodeToString(bytesContent)

		tags := nostr.Tags{}
		// 置き換え可能なイベントの場合
		if replaceable {
			fileIdentifier := getReplaceableIdentifier(indexHtmlIdentifier, filePath)
			tags = tags.AppendUnique(nostr.Tag{"d", fileIdentifier})
		}

		// 拡張子からContent-Typeを取得
		contentType := availableTextFileContentTypes[filepath.Ext(filePath)]
		tags = tags.AppendUnique(nostr.Tag{"type", contentType})

		// kindを設定
		var kind int
		if replaceable {
			kind = 30064
		} else {
			kind = 1064
		}

		// eventを取得
		event, err := getEvent(priKey, pubKey, content, kind, tags)
		if err != nil {
			return err
		}

		textFilePathToEvent[filePath] = event

		addNostrEventQueue(event, filePath)
	}

	return nil
}
