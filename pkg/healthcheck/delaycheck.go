package healthcheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/ssrlive/proxypool/log"
	"github.com/ssrlive/proxypool/pkg/proxy"

	"github.com/ivpusic/grpool"

	"github.com/Dreamacro/clash/adapter"
)

func CleanBadProxiesWithGrpool(proxies []proxy.Proxy) (cproxies []proxy.Proxy) {
	// Note: Grpool实现对go并发管理的封装，主要是在数据量大时减少内存占用，不会提高效率。
	log.Debugln("[delaycheck.go] connection: %d, timeout: %.2fs", DelayConn, DelayTimeout.Seconds())
	numWorker := DelayConn
	numJob := 1
	if numWorker > 4 {
		numJob = (numWorker + 2) / 3
	}
	pool := grpool.NewPool(numWorker, numJob)
	cproxies = make(proxy.ProxyList, 0, 500)

	m := sync.Mutex{}

	pool.WaitCount(len(proxies))
	doneCount := 0
	dcm := sync.Mutex{}
	// 线程：延迟测试，测试过程通过grpool的job并发
	go func() {
		for _, p := range proxies {
			pp := p // 捕获，否则job执行时是按当前的p测试的
			pool.JobQueue <- func() {
				defer pool.JobDone()
				delay, err := testDelay(pp)
				if err == nil && delay != 0 {
					m.Lock()
					cproxies = append(cproxies, pp)
					if ps, ok := ProxyStats.Find(pp); ok {
						ps.UpdatePSDelay(delay)
					} else {
						ps = &Stat{
							Id:    pp.Identifier(),
							Delay: delay,
						}
						ProxyStats = append(ProxyStats, *ps)
					}
					m.Unlock()
				}
				// Progress status
				dcm.Lock()
				doneCount++
				progress := float64(doneCount) * 100 / float64(len(proxies))
				fmt.Printf("\r\t[%5.1f%% DONE]", progress)
				dcm.Unlock()
			}
		}
	}()

	pool.WaitAll()
	pool.Release()
	fmt.Println()
	return
}

// Return 0 for error
func testDelay(p proxy.Proxy) (delay time.Duration, err error) {
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return
	}
	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
	}

	if proxy.GoodNodeThatClashUnsupported(p) {
		host := pmap["server"].(string)
		port := fmt.Sprint(pmap["port"].(int))
		if _, interval, err := netConnectivity(host, port); err == nil {
			return interval, nil
		} else {
			return 0, err
		}
	}

	clashProxy, err := adapter.ParseProxy(pmap)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	respC := make(chan struct {
		time.Duration
		error
	})
	defer close(respC)
	go func() {
		sTime := time.Now()
		err = HTTPHeadViaProxy(clashProxy, "http://www.gstatic.com/generate_204")
		respC <- struct {
			time.Duration
			error
		}{time.Since(sTime), err}
	}()

	pair, ok := <-respC

	if ok {
		return pair.Duration, pair.error
	} else {
		return 0, context.DeadlineExceeded
	}
}

func netConnectivity(host string, port string) (string, time.Duration, error) {
	result := ""
	timeout := time.Second * 3
	beginning := time.Now()
	interval := timeout
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if conn != nil {
		defer conn.Close()
		result, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
		interval = time.Since(beginning)
	}
	return result, interval, err
}

func CleanBadProxies(proxies []proxy.Proxy) (cproxies []proxy.Proxy) {
	cproxies = make(proxy.ProxyList, 0, 500)
	for _, p := range proxies {
		delay, err := testDelay(p)
		if err == nil && delay != 0 {
			cproxies = append(cproxies, p)
			if ps, ok := ProxyStats.Find(p); ok {
				ps.UpdatePSDelay(delay)
			} else {
				ps = &Stat{
					Id:    p.Identifier(),
					Delay: delay,
				}
				ProxyStats = append(ProxyStats, *ps)
			}
		}
	}
	return
}

func PingFromChina(host string, port string) (bool, time.Duration, error) {
	beginning := time.Now()

	url1 := "https://tool.chinaz.com/port?host=" + host + "&port=" + port
	doc, err := htmlquery.LoadURL(url1)
	if err != nil {
		return false, 0, err
	}
	title, err := htmlquery.Query(doc, "//title")
	if err != nil {
		return false, 0, err
	}
	if title.FirstChild.Data != host+"网站端口扫描结果" {
		return false, 0, errors.New("title not match")
	}
	inputs, err := htmlquery.QueryAll(doc, "//input")
	if err != nil {
		return false, 0, err
	}
	var encoded string
	var found bool
	for _, input := range inputs {
		for _, a := range input.Attr {
			if a.Key == "id" && a.Val == "encode" {
				found = true
				break
			}
		}
		if found {
			for _, a := range input.Attr {
				if a.Key == "value" {
					encoded = a.Val
					break
				}
			}
			break
		}
	}
	if encoded == "" {
		return false, 0, errors.New("encode not found")
	}

	url2 := "https://tool.chinaz.com/iframe.ashx?t=port"
	resp, err := http.PostForm(url2, url.Values{"host": {host}, "port": {port}, "encode": {encoded}})
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, 0, err
	}

	result := strings.Contains(string(body), "status:1")

	interval := time.Since(beginning)

	return result, interval, nil
}
