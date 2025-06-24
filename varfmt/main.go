package main

import (
	"fmt"
	"strconv"
	"strings"
)

func Sprintf(format string, args ...interface{}) string {
	var res strings.Builder
	cnt := 0
	for i := 0; i < len(format); i++ {
		if format[i] == '{' {
			i++
			numS := ""
			for ; format[i] != '}'; i++ {
				numS += string(format[i])
			}
			if numS != "" {
				num, _ := strconv.Atoi(numS)
				res.WriteString(fmt.Sprint(args[num]))
			} else {
				res.WriteString(fmt.Sprint(args[cnt]))
			}
			cnt++
		} else {
			res.WriteString(string(format[i]))
		}
	}
	return res.String()
}
