package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	C "github.com/Dreamacro/clash/constant"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func DISNEYTest(p C.Proxy) (sCode int, err error, resultBody string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url := "https://disney.api.edge.bamgrid.com/devices"
	url2 := "https://disney.api.edge.bamgrid.com/token"
	url3 := "https://disney.api.edge.bamgrid.com/graph/v1/device/graphql"
	body := `{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","attributes":{}}`
	UA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36"
	headers := map[string]string{
		"User-Agent":      UA,
		"content-type":    "application/json; charset=UTF-8",
		"authorization":   "Bearer ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
		"accept-encoding": "gzip, deflate, br",
	}
	sCode, err, resultBody = connectTest(url, p, ctx, body, http.MethodPost, headers)
	if resultBody != "" {
		var preAssertionJSON map[string]interface{}
		err = json.Unmarshal([]byte(resultBody), &preAssertionJSON)
		if err == nil {
			if assertion, ok := preAssertionJSON["assertion"].(string); ok {
				disneyCookie := fmt.Sprintf("grant_type=urn%%3Aietf%%3Aparams%%3Aoauth%%3Agrant-type%%3Atoken-exchange&latitude=0&longitude=0&platform=browser&subject_token=%s&subject_token_type=urn%%3Abamtech%%3Aparams%%3Aoauth%%3Atoken-type%%3Adevice", assertion)
				headers = map[string]string{
					"User-Agent":    UA,
					"authorization": "Bearer ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
				}
				sCode, err, resultBody = connectTest(url2, p, ctx, disneyCookie, http.MethodPost, headers)
				if resultBody != "" {
					var tokenJSON map[string]interface{}
					err = json.Unmarshal([]byte(resultBody), &tokenJSON)
					if err == nil {
						if errorDescription, ok := tokenJSON["error_description"].(string); ok {
							if errorDescription != "" {
								err = errors.New(errorDescription)
								return
							}
						}
					}
				}
				body = `{"query":"mutation registerDevice($input: RegisterDeviceInput!) { registerDevice(registerDevice: $input) { grant { grantType assertion } } }","variables":{"input":{"deviceFamily":"browser","applicationRuntime":"chrome","deviceProfile":"windows","deviceLanguage":"en","attributes":{"osDeviceIds":[],"manufacturer":"microsoft","model":null,"operatingSystem":"windows","operatingSystemVersion":"10.0","browserName":"chrome","browserVersion":"96.0.4606"}}}}`
				headers = map[string]string{
					"User-Agent":      UA,
					"Content-Type":    "application/json",
					"Accept-Language": "en",
					"authorization":   "ZGlzbmV5JmJyb3dzZXImMS4wLjA.Cu56AgSfBTDag5NiRA81oLHkDZfu5L3CKadnefEAY84",
				}
				sCode, err, resultBody = connectTest(url3, p, ctx, body, http.MethodPost, headers)
				if resultBody != "" {
					var dataJSON map[string]interface{}
					err = json.Unmarshal([]byte(resultBody), &dataJSON)
					if err == nil {
						err = errors.New("get country Code fail")
						if extensions, ok := dataJSON["extensions"].(map[string]interface{}); ok {
							if sdk, ok := extensions["sdk"].(map[string]interface{}); ok {
								if session, ok := sdk["session"].(map[string]interface{}); ok {
									if countryCode, ok := session["location"].(map[string]interface{})["countryCode"].(string); ok {
										if countryCode != "" {
											resultBody = countryCode
											err = nil
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return
}

func connectTest(url string, p C.Proxy, ctx context.Context, body string, method string, headers map[string]string) (sCode int, err error, resultBody string) {
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

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
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
	for i := 0; i < 3; i++ {
		response, error2 := client.Do(req)
		err = error2
		if error2 != nil {
			continue
		}
		defer response.Body.Close()
		responseBody, error2 := io.ReadAll(response.Body)
		err = error2
		if error2 != nil {
			continue
		}
		resultBody = string(responseBody)
		sCode = response.StatusCode
		break
	}
	return
}
