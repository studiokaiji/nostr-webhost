package tools

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr/nip19"
)

func ResolvePubKey(npubOrHex string) (string, error) {
	// npubから始まる場合はデコードする
	if npubOrHex[0:4] == "npub" {
		_, v, err := nip19.Decode(npubOrHex)
		if err != nil {
			return "", fmt.Errorf("Invalid npub")
		}
		return v.(string), nil
	} else {
		_, err := nip19.EncodePublicKey(npubOrHex)
		if err != nil {
			return "", fmt.Errorf("Invalid pubkey")
		}
	}
	return npubOrHex, nil
}
