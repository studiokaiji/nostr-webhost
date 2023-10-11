package server

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/relays"
)

func Start(port string) {
	ctx := context.Background()

	allRelays, err := relays.GetAllRelays()
	if err != nil {
		panic(err)
	}

	pool := nostr.NewSimplePool(ctx)

	r := gin.Default()

	r.GET("/e/:hex_or_nevent", func(ctx *gin.Context) {
		hexOrNevent := ctx.Param("hex_or_nevent")

		ids := []string{}

		// neventからIDを取得
		if hexOrNevent[0:6] == "nevent" {
			_, res, err := nip19.Decode(hexOrNevent)
			if err != nil {
				ctx.String(http.StatusBadRequest, "Invalid nevent")
				return
			}

			data, ok := res.(nostr.EventPointer)
			if !ok {
				ctx.String(http.StatusBadRequest, "Failed to decode nevent")
				return
			}

			ids = append(ids, data.ID)
			allRelays = append(allRelays, data.Relays...)
		} else {
			ids = append(ids, hexOrNevent)
		}

		// Poolからデータを取得する
		ev := pool.QuerySingle(ctx, allRelays, nostr.Filter{
			Kinds: []int{consts.KindWebhostHTML, consts.KindWebhostCSS, consts.KindWebhostJS, consts.KindWebhostPicture},
			IDs:   ids,
		})
		if ev != nil {
			switch ev.Kind {
			case consts.KindWebhostHTML:
				ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(ev.Content))
			case consts.KindWebhostCSS:
				ctx.Data(http.StatusOK, "text/css; charset=utf-8", []byte(ev.Content))
			case consts.KindWebhostJS:
				ctx.Data(http.StatusOK, "text/javascript; charset=utf-8", []byte(ev.Content))
			case consts.KindWebhostPicture:
				{
					eTag := ev.Tags.GetFirst([]string{"e"})
					mTag := ev.Tags.GetFirst([]string{"m"})

					if eTag == nil || mTag == nil {
						ctx.String(http.StatusBadRequest, http.StatusText((http.StatusBadRequest)))
						return
					}

					evData := pool.QuerySingle(ctx, allRelays, nostr.Filter{
						IDs: []string{eTag.Value()},
					})

					if evData == nil {
						ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
						return
					}

					data, err := base64.StdEncoding.DecodeString(evData.Content)
					if err != nil {
						ctx.String(http.StatusBadRequest, http.StatusText((http.StatusBadRequest)))
						return
					}

					ctx.Data(http.StatusOK, mTag.Value(), data)
				}
			default:
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			}
		} else {
			ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		return
	})

	// Replaceable Event (NIP-33)
	r.GET("/p/:pubKey/d/*dTag", func(ctx *gin.Context) {
		// pubKeyを取得しFilterに追加
		pubKey := ctx.Param("pubKey")
		// npubから始まる場合はデコードする
		if pubKey[0:4] == "npub" {
			_, v, err := nip19.Decode(pubKey)
			if err != nil {
				ctx.String(http.StatusBadRequest, "Invalid npub")
				return
			}
			pubKey = v.(string)
		}
		authors := []string{pubKey}

		// dTagを取得しFilterに追加
		// dTagの最初は`/`ではじまるのでそれをslice
		dTag := ctx.Param("dTag")[1:]

		tags := nostr.TagMap{}
		tags["d"] = []string{dTag}

		// Poolからデータを取得する
		ev := pool.QuerySingle(ctx, allRelays, nostr.Filter{
			Kinds: []int{
				consts.KindWebhostReplaceableHTML,
				consts.KindWebhostReplaceableCSS,
				consts.KindWebhostReplaceableJS,
			},
			Authors: authors,
			Tags:    tags,
		})
		if ev != nil {
			switch ev.Kind {
			case consts.KindWebhostReplaceableHTML:
				ctx.Data(http.StatusOK, "text/html; charset=utf-8", []byte(ev.Content))
			case consts.KindWebhostReplaceableCSS:
				ctx.Data(http.StatusOK, "text/css; charset=utf-8", []byte(ev.Content))
			case consts.KindWebhostReplaceableJS:
				ctx.Data(http.StatusOK, "text/javascript; charset=utf-8", []byte(ev.Content))
			default:
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			}
		} else {
			ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		return
	})

	r.Run(":" + port)
}
