package paths

import (
	"os"
	"path/filepath"
)

const BaseDirName = ".nostr-webhost"

func GetSettingsDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dirPath := filepath.Join(homeDir, BaseDirName)
	if os.IsNotExist(err) {
		// ディレクトリが存在しない場合に作成
		err = os.Mkdir(dirPath, 0700)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	return dirPath, nil
}
