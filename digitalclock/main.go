package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	port := flag.String("port", "", "port to listen on")
	flag.Parse()
	if *port == "" {
		fmt.Fprintln(os.Stderr, "port is required")
		os.Exit(1)
	}

	http.HandleFunc("/", clockHandler)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintln(os.Stderr, "server failed:", err)
		os.Exit(1)
	}
}

func clockHandler(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	k := 1
	if ks := q.Get("k"); ks != "" {
		ki, err := strconv.Atoi(ks)
		if err != nil || ki < 1 || ki > 30 {
			http.Error(rw, "invalid k", http.StatusBadRequest)
			return
		}
		k = ki
	}

	timeStr := q.Get("time")
	if timeStr == "" {
		timeStr = time.Now().Format("15:04:05")
	} else {
		if !validTimeFormat(timeStr) {
			http.Error(rw, "invalid time", http.StatusBadRequest)
			return
		}
	}

	symbols := []string{}
	for _, ch := range timeStr {
		sym := bitmapFor(ch)
		if sym == "" {
			http.Error(rw, "invalid time", http.StatusBadRequest)
			return
		}
		symbols = append(symbols, sym)
	}

	first := strings.Split(symbols[0], "\n")

	width := (6*len(first[0]) + 2*len(strings.Split(Colon, "\n")[0])) * k
	height := len(first) * k

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.White)
		}
	}

	offX := 0
	for _, sym := range symbols {
		block := strings.Split(sym, "\n")
		for sy, row := range block {
			for sx, pixel := range row {
				if pixel != '.' {
					for dy := 0; dy < k; dy++ {
						for dx := 0; dx < k; dx++ {
							img.Set(offX+sx*k+dx, sy*k+dy, Cyan)
						}
					}
				}
			}
		}
		offX += len(block[0]) * k
	}

	rw.Header().Set("Content-Type", "image/png")
	rw.WriteHeader(http.StatusOK)
	_ = png.Encode(rw, img)
}

func validTimeFormat(s string) bool {
	if len(s) != 8 {
		return false
	}
	if s[2] != ':' || s[5] != ':' {
		return false
	}

	h, err := strconv.Atoi(s[0:2])
	if err != nil || h < 0 || h > 23 {
		return false
	}

	m, err := strconv.Atoi(s[3:5])
	if err != nil || m < 0 || m > 59 {
		return false
	}

	sec, err := strconv.Atoi(s[6:8])
	if err != nil || sec < 0 || sec > 59 {
		return false
	}
	return true
}

func bitmapFor(ch rune) string {
	switch ch {
	case '0':
		return Zero
	case '1':
		return One
	case '2':
		return Two
	case '3':
		return Three
	case '4':
		return Four
	case '5':
		return Five
	case '6':
		return Six
	case '7':
		return Seven
	case '8':
		return Eight
	case '9':
		return Nine
	case ':':
		return Colon
	default:
		return ""
	}
}
