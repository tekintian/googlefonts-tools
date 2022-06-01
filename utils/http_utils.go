package utils

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// http get request with user defined ip and ua
func HttpGet(url string, ip, ua string) ([]byte, error) {
	client := &http.Client{
		Timeout: 500 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("X-FORWARDED-FOR", ip)
	req.Header.Set("CLIENT-IP", ip)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, err
}

// GetClientIp returns the client ip of this request without port.
// Note that this ip address might be modified by client header.
func GetClientIp(r *http.Request) string {
	clientIp := ""
	realIps := r.Header.Get("X-Forwarded-For")
	if realIps != "" && len(realIps) != 0 && !strings.EqualFold("unknown", realIps) {
		ipArray := strings.Split(realIps, ",")
		clientIp = ipArray[0]
	}
	if clientIp == "" {
		clientIp = r.Header.Get("Proxy-Client-IP")
	}
	if clientIp == "" {
		clientIp = r.Header.Get("WL-Proxy-Client-IP")
	}
	if clientIp == "" {
		clientIp = r.Header.Get("HTTP_CLIENT_IP")
	}
	if clientIp == "" {
		clientIp = r.Header.Get("HTTP_X_FORWARDED_FOR")
	}
	if clientIp == "" {
		clientIp = r.Header.Get("X-Real-IP")
	}
	if clientIp == "" {
		clientIp = GetRemoteIp(r)
	}
	return clientIp
}

// GetRemoteIp returns the ip from RemoteAddr.
func GetRemoteIp(r *http.Request) string {
	regex, err := regexp.Compile(`(.+):(\d+)`)
	if err != nil {
		return ""
	}
	array := regex.FindStringSubmatch(r.RemoteAddr)
	if len(array) > 1 {
		return strings.Trim(array[1], "[]")
	}
	return r.RemoteAddr
}
