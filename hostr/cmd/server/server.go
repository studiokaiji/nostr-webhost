package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/relays"
	"github.com/studiokaiji/nostr-webhost/hostr/cmd/tools"
)

func Start(port string, mode string) {
	ctx := context.Background()

	allRelays, err := relays.GetAllRelays()
	if err != nil {
		panic(err)
	}

	pool := nostr.NewSimplePool(ctx)

	r := gin.Default()

	r.GET("/e/:hex_or_nevent", func(ctx *gin.Context) {
		hexOrNevent := ctx.Param("hex_or_nevent")

		subdomainPubKey := ""

		if mode == "secure" {
			// modeがsecureの場合、サブドメインにnpubが含まれていないルーティングは許可しない
			host := ctx.Request.Host
			subdomain := strings.Split(host, ".")[0]
			subdomainPubKey, err = tools.ResolvePubKey(subdomain)
			if err != nil {
				ctx.String(http.StatusBadRequest, "Routing without npub in the subdomain is not allowed")
				return
			}
		}

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

		filter := nostr.Filter{
			Kinds: []int{consts.KindWebhostHTML, consts.KindWebhostCSS, consts.KindWebhostJS, consts.KindWebhostPicture},
			IDs:   ids,
		}
		if mode == "secure" {
			filter.Authors = []string{subdomainPubKey}
		}

		// Poolからデータを取得する
		ev := pool.QuerySingle(ctx, allRelays, filter)
		if ev != nil {
			contentType, err := tools.GetContentType(ev.Kind)
			if err != nil {
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			}
			ctx.Data(http.StatusOK, contentType, []byte(ev.Content))
		} else {
			ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		return
	})

	if mode != "secure" {
		r.GET("/p/:pubKey/d/*dTag", func(ctx *gin.Context) {
			// pubKeyを取得しFilterに追加
			pubKey := ctx.Param("pubKey")
			pubKey, err := tools.ResolvePubKey(pubKey)
			if err != nil {
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
				return
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
				contentType, err := tools.GetContentType(ev.Kind)
				if err != nil {
					ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
				}
				ctx.Data(http.StatusOK, contentType, []byte(ev.Content))
			} else {
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			}

			return
		})

	}

	if mode != "normal" {
		r.GET("/d/*dTag", func(ctx *gin.Context) {
			host := ctx.Request.Host
			subdomain := strings.Split(host, ".")[0]

			// subdomainからpubKeyを取得しFilterに追加
			pubKey, err := tools.ResolvePubKey(subdomain)
			if err != nil {
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
				return
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
				contentType, err := tools.GetContentType(ev.Kind)
				if err != nil {
					ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
				}
				ctx.Data(http.StatusOK, contentType, []byte(ev.Content))
			} else {
				ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
			}

			return
		})
	}

	r.Run(":" + port)
}
