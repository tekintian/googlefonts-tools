//go:build ignore

package main

import (
	"fmt"
	"regexp"
	"strings"
)

func GenerateSelfHostCSS(cssContent string) string {
	urlReg := regexp.MustCompile(`url\((https://fonts\.gstatic\.com/s/[^)]+)\)`)
	ffReg := regexp.MustCompile(`https://fonts\.gstatic\.com/s/([^/]+)/([^/]+)/(.+)`)

	result := urlReg.ReplaceAllStringFunc(cssContent, func(urlStr string) string {
		submatch := urlReg.FindStringSubmatch(urlStr)
		if len(submatch) < 2 {
			return urlStr
		}
		matches := ffReg.FindStringSubmatch(submatch[1])
		if len(matches) >= 4 {
			relPath := fmt.Sprintf("../%s/%s", matches[2], matches[3])
			return "url(" + relPath + ")"
		}
		return urlStr
	})

	return result
}

func main() {
	input := `/* latin-ext */
@font-face {
  font-family: 'Inter';
  font-style: normal;
  font-weight: 400;
  font-display: swap;
  src: url(https://fonts.gstatic.com/s/inter/v18/UcCO3FwrK3iLTeHuSnVMgXe5R3-5Dw.woff2) format('woff2');
  unicode-range: U+0100-02AF, U+0304, U+0308, U+0329;
}
/* latin */
@font-face {
  font-family: 'Inter';
  font-style: normal;
  font-weight: 700;
  font-display: swap;
  src: url(https://fonts.gstatic.com/s/inter/v18/UcCO3FwrK3iLTeHuSfVMgXe5R3-5Dw.woff2) format('woff2');
  unicode-range: U+0000-00FF, U+0131, U+0152-0153;
}`

	output := GenerateSelfHostCSS(input)

	fmt.Println("=== Self-Host CSS Output ===")
	fmt.Println(output)
	fmt.Println()

	pass := true

	if strings.Contains(output, "fonts.gstatic.com") {
		fmt.Println("FAIL: Still contains fonts.gstatic.com URLs")
		pass = false
	} else {
		fmt.Println("PASS: All gstatic URLs replaced with relative paths")
	}

	if strings.Contains(output, "../v18/") {
		fmt.Println("PASS: Relative font paths correctly generated")
	} else {
		fmt.Println("FAIL: Relative font paths not found")
		pass = false
	}

	if strings.Contains(output, "@font-face") {
		fmt.Println("PASS: @font-face rules preserved")
	} else {
		fmt.Println("FAIL: @font-face rules not preserved")
		pass = false
	}

	if strings.Contains(output, "))") {
		fmt.Println("FAIL: Double closing parenthesis detected")
		pass = false
	} else {
		fmt.Println("PASS: No double closing parenthesis")
	}

	if strings.Contains(output, "format('woff2')") {
		fmt.Println("PASS: format() preserved correctly")
	} else {
		fmt.Println("FAIL: format() not preserved")
		pass = false
	}

	expectedPath := "url(../v18/UcCO3FwrK3iLTeHuSnVMgXe5R3-5Dw.woff2)"
	if strings.Contains(output, expectedPath) {
		fmt.Println("PASS: Exact font path matches expected")
	} else {
		fmt.Println("FAIL: Font path doesn't match expected")
		pass = false
	}

	if pass {
		fmt.Println("\n=== ALL TESTS PASSED ===")
	} else {
		fmt.Println("\n=== SOME TESTS FAILED ===")
	}
}
