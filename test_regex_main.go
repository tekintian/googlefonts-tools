//go:build ignore

package main

import (
	"fmt"
	"regexp"
)

func main() {
	css := `src: url(https://fonts.gstatic.com/s/inter/v18/UcCO3FwrK3iLTeHuSnVMgXe5R3-5Dw.woff2) format('woff2');`

	urlReg := regexp.MustCompile(`url\((https://fonts\.gstatic\.com/s/[^)]+)\)`)
	ffReg := regexp.MustCompile(`https://fonts\.gstatic\.com/s/([^/]+)/([^/]+)/(.*)`)

	result := urlReg.ReplaceAllStringFunc(css, func(urlStr string) string {
		matches := ffReg.FindStringSubmatch(urlStr)
		fmt.Printf("Callback input: [%s]\n", urlStr)
		fmt.Printf("ffReg matches: %v (len=%d)\n", matches, len(matches))
		if len(matches) >= 4 {
			relPath := fmt.Sprintf("./fonts/%s/%s/%s", matches[1], matches[2], matches[3])
			result := "url(" + relPath + ")"
			fmt.Printf("Callback output: [%s]\n", result)
			return result
		}
		return urlStr
	})

	fmt.Printf("Result: [%s]\n", result)
}
