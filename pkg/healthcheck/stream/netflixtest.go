package stream

import (
	"context"
	C "github.com/Dreamacro/clash/constant"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"
)

func NETFLIXTest(p C.Proxy, url string) (sCode int, err error, country string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addr, err := urlToMetadata(url)
	if err != nil {
		return
	}
	instance, err := p.DialContext(ctx, &addr)
	if err != nil {
		return
	}
	defer func(instance C.Conn) {
		err := instance.Close()
		if err != nil {
			return
		}
	}(instance)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req = req.WithContext(ctx)

	transport := &http.Transport{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			return instance, nil
		},
		// from http.DefaultTransport
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := http.Client{
		Transport: transport,
	}
	defer client.CloseIdleConnections()

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	//err = resp.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	bodyString := string(bodyBytes)
	re := regexp.MustCompile(`"country":"(\w+)"`)
	match := re.FindStringSubmatch(bodyString)
	if len(match) > 1 {
		result := match[1]
		country = result
	}
	if err != nil {
		return
	}
	sCode = resp.StatusCode
	return
}
