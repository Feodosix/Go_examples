package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	for _, url := range os.Args[1:] {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("fetch: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("fetch: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(body))
	}
}
