package keystore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/paths"
)

const PATH = ".nostr_account_secret"

func SetSecret(key string) error {
	// nsecから始まる場合はデコードする
	if key[0:4] == "nsec" {
		_, v, err := nip19.Decode(key)
		if err != nil {
			panic(err)
		}
		key = v.(string)
	}

	dir, err := paths.GetSettingsDirectory()
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, PATH)
	if err != nil {
		return err
	}

	// キーをファイルに書き込み
	return os.WriteFile(filePath, []byte(key), 0644)
}

func ShowPublic() (string, string, error) {
	hex, err := GetPublic()
	if err != nil {
		return "", "", err
	}
	npub, err := nip19.EncodePublicKey(hex)
	fmt.Printf("npub: %s\nhex: %s\n", npub, hex)
	return hex, npub, nil
}

func GetPublic() (string, error) {
	secret, err := GetSecret()
	if err != nil {
		return "", err
	}
	hex, err := nostr.GetPublicKey(secret)
	return hex, err
}

func GetSecret() (string, error) {
	dir, err := paths.GetSettingsDirectory()
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(dir, PATH)
	if err != nil {
		return "", err
	}

	secretBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", errors.New("Could not read secret")
	}
	return string(secretBytes), nil
}
