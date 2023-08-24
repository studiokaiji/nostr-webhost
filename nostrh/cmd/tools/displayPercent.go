package tools

import (
	"fmt"
	"time"

	"golang.org/x/term"
)

func DisplayProgressBar(current, total *int) {
	// ターミナルのサイズを取得
	terminalWidth, _, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}

	width := terminalWidth - 12

	// ターミナルの幅を最大100として調整
	if width > 100 {
		width = 100
	}

	for {
		progress := int(float64(*current) / float64(*total) * float64(width))

		// バーの描画
		bar := ""
		for j := 0; j < width; j++ {
			if j < progress {
				bar += "="
			} else {
				bar += " "
			}
		}

		// カーソルを行の先頭に戻して上書き
		fmt.Printf("\r[%s] %d/%d", bar, *current, *total)

		if *current >= *total {
			fmt.Println("")
			break
		}

		time.Sleep(20 * time.Millisecond)
	}
}
