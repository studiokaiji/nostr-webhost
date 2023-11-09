package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 特定のパス以下のファイルを検索し、与えられたsuffixesに該当するファイルのパスのみを返す
func FindFilesWithBasePathBySuffixes(basePath string, suffixes []string) ([]string, error) {
	filePaths := []string{}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// ディレクトリはスキップ
		if !info.IsDir() {
			// 各サフィックスに対してマッチングを試みる
			for _, suffix := range suffixes {
				// ファイル名とサフィックスがマッチした場合
				if strings.HasSuffix(strings.ToLower(info.Name()), strings.ToLower(suffix)) {
					// フルパスからbasePathまでの相対パスを計算
					if err != nil {
						fmt.Println("❌ Error calculating relative path:", err)
						continue
					}
					// マッチするファイルの相対パスをスライスに追加
					filePaths = append(filePaths, path)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}
