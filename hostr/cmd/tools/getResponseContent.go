package tools

import (
	"encoding/base64"
)

func GetResponseContent(eventContent string, isTextFile bool) ([]byte, error) {
	if isTextFile {
		// NIP-95ファイル(Text File)の場合はbase64エンコードされているのでdecodeする
		return base64.StdEncoding.DecodeString(eventContent)
	} else {
		return []byte(eventContent), nil
	}
}
