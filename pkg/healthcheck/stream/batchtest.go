package stream

import (
	"fmt"
	C "github.com/Dreamacro/clash/constant"
	"github.com/ivpusic/grpool"
	"sync"
)

// BatchCheck : n int, to set ConcurrencyNum.
func BatchCheck(proxiesList []C.Proxy, n int) (NETFLIXList []Element) {
	//  Grpool实现对go并发管理的封装
	numWorker := n
	numJob := 1
	if numWorker > 4 {
		numJob = (numWorker + 2) / 3
	}
	pool := grpool.NewPool(numWorker, numJob)
	m := sync.Mutex{}
	pool.WaitCount(len(proxiesList))
	doneCount := 0
	dcm := sync.Mutex{}
	go func() {
		for _, p := range proxiesList {
			pp := p
			pool.JobQueue <- func() {
				defer pool.JobDone()
				sCode, err, country := NETFLIXTest(pp, "https://www.netflix.com/title/70143836")
				if err == nil && sCode == 200 {
					m.Lock()
					NETFLIXList = append(NETFLIXList, Element{Name: pp.Name(), Country: country})
					m.Unlock()
				}
				// Progress status
				dcm.Lock()
				doneCount++
				progress := float64(doneCount) * 100 / float64(len(proxiesList))
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

func BatchDisneyCheck(proxiesList []C.Proxy, n int) (DISNEYList []Element) {
	//  Grpool实现对go并发管理的封装
	numWorker := n
	numJob := 1
	if numWorker > 4 {
		numJob = (numWorker + 2) / 3
	}
	pool := grpool.NewPool(numWorker, numJob)
	m := sync.Mutex{}
	pool.WaitCount(len(proxiesList))
	doneCount := 0
	dcm := sync.Mutex{}
	go func() {
		for _, p := range proxiesList {
			pp := p
			pool.JobQueue <- func() {
				defer pool.JobDone()
				sCode, err, country := DISNEYTest(pp)
				if err == nil && sCode == 200 {
					m.Lock()
					DISNEYList = append(DISNEYList, Element{Name: pp.Name(), Country: country})
					m.Unlock()
				}
				// Progress status
				dcm.Lock()
				doneCount++
				progress := float64(doneCount) * 100 / float64(len(proxiesList))
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
