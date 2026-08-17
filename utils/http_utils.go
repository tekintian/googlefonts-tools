package utils

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func HttpGet(url string, ip, ua string) ([]byte, error) {
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if ua == "" {
		ua = DefaultUserAgent
	}
	req.Header.Set("User-Agent", ua)
	if ip != "" {
		req.Header.Set("X-FORWARDED-FOR", ip)
		req.Header.Set("CLIENT-IP", ip)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetFakeIp() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i1 := r.Intn(250)
	if i1 == 127 || i1 == 10 || i1 == 192 || i1 == 172 || i1 == 0 {
		i1 = 100 + r.Intn(150)
	}
	return fmt.Sprintf("%d.%d.%d.%d", i1, r.Intn(250), r.Intn(250), r.Intn(250))
}
