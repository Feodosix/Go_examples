package main

import "strings"

var ones = []string{"", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine"}
var teens = []string{"ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
var tens = []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
var thousands = []string{"", "thousand", "million", "billion"}

func Spell(n int64) string {
	if n == 0 {
		return "zero"
	}

	var res strings.Builder
	if n < 0 {
		res.WriteString("minus ")
		n = -n
	}

	var parts []string
	for i := 0; n > 0; i++ {
		if n%1000 != 0 {
			parts = append([]string{spellHundreds(int(n%1000)) + " " + thousands[i]}, parts...)
		}
		n /= 1000
	}

	res.WriteString(strings.TrimSpace(strings.Join(parts, " ")))
	return res.String()
}

func spellHundreds(n int) string {
	var res []string
	if n >= 100 {
		res = append(res, ones[n/100]+" hundred")
		n %= 100
	}
	if n >= 20 {
		res = append(res, tens[n/10])
		n %= 10
		if n > 0 {
			res[len(res)-1] += "-" + ones[n]
			n = 0
		}
	}
	if n >= 10 {
		res = append(res, teens[n-10])
		n = 0
	}
	if n > 0 {
		res = append(res, ones[n])
	}
	return strings.Join(res, " ")
}
