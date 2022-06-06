package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"sync"
	"time"

	"github.com/ivpusic/grpool"

	"github.com/Dreamacro/clash/adapters/outbound"
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
func testDelay(p proxy.Proxy) (delay uint16, err error) {
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return
	}
	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
	}

	clashProxy, err := outbound.ParseProxy(pmap)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}

	// Custom context time to avoid unexpected connection block due to dependency
	respC := make(chan uint16)
	m := sync.Mutex{}
	closed := false
	defer close(respC)
	go func() {
		sTime := time.Now()
		err = HTTPHeadViaProxy(clashProxy, "http://www.gstatic.com/generate_204")
		m.Lock()
		if closed {
			m.Unlock()
			return
		}
		m.Unlock()
		if err != nil {
			respC <- 0
			return
		}
		fTime := time.Now()
		d := uint16(fTime.Sub(sTime) / time.Millisecond)
		respC <- d
	}()

	select {
	case delay = <-respC:
		m.Lock()
		closed = true
		m.Unlock()
		return delay, nil
	case <-time.After(DelayTimeout * 2):
		log.Debugln("unexpected delay check timeout error in proxy %s\n", p.Link())
		m.Lock()
		closed = true
		m.Unlock()
		return 0, context.DeadlineExceeded
	}
}
