package server

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nbd-wtf/go-nostr"
	"github.com/studiokaiji/nostr-webhost/nostrh/cmd/consts"
	"github.com/studiokaiji/nostr-webhost/nostrh/cmd/relays"
)

func Start(port string) {
	ctx := context.Background()

	allRelays, err := relays.GetAllRelays()
	if err != nil {
		panic(err)
	}

	pool := nostr.NewSimplePool(ctx)

	r := gin.Default()

	r.GET("/e/:idHex", func(ctx *gin.Context) {
		// IDを取得
		id := ctx.Param("idHex")

		// Poolからデータを取得する
		ev := pool.QuerySingle(ctx, allRelays, nostr.Filter{
			Kinds: []int{consts.KindWebhostHTML, consts.KindWebhostCSS, consts.KindWebhostJS, consts.KindWebhostPicture},
			IDs:   []string{id},
		})
		if ev != nil {
			switch ev.Kind {
			case consts.KindWebhostHTML:
				ctx.Data(http.StatusOK, "text/html", []byte(ev.Content))
			case consts.KindWebhostCSS:
				ctx.Data(http.StatusOK, "text/css", []byte(ev.Content))
			case consts.KindWebhostJS:
				ctx.Data(http.StatusOK, "text/javascript", []byte(ev.Content))
			case consts.KindWebhostPicture:
				{
					eTag := ev.Tags.GetFirst([]string{"e"})
					mTag := ev.Tags.GetFirst([]string{"e"})

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
		}

		ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	})

	r.Run(":" + port)
}
