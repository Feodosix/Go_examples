package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	counts := make(map[string]int)
	for _, arg := range args {
		data, err := os.ReadFile(arg)
		if err != nil {
			panic(err)
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			counts[line]++
		}
	}
	for line, count := range counts {
		if count > 1 {
			fmt.Printf("%v\t%v\n", count, line)
		}
	}
}
