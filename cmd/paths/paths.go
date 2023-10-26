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

	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		// ディレクトリが存在しない場合に作成
		err = os.Mkdir(dirPath, 0700)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", nil
	}

	return dirPath, nil
}

func GetProjectRootDirectory() (string, error) {
	// 実行中のバイナリの絶対パスを取得
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	// ディレクトリパスを取得
	dir := filepath.Dir(exePath)
	return dir, nil
}
