package app

import (
	"fmt"
	C "github.com/Sansui233/proxypool/config"
	"github.com/Sansui233/proxypool/internal/cache"
	"github.com/Sansui233/proxypool/internal/database"
	"github.com/Sansui233/proxypool/log"
	"github.com/Sansui233/proxypool/pkg/geoIp"
	"github.com/Sansui233/proxypool/pkg/healthcheck"
	"github.com/Sansui233/proxypool/pkg/provider"
	"github.com/Sansui233/proxypool/pkg/proxy"
	"sync"
	"time"
)

var location, _ = time.LoadLocation("PRC")

func CrawlGo() {
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)
	for _, g := range Getters {
		wg.Add(1)
		go g.Get2ChanWG(pc, wg)
	}
	proxies := cache.GetProxies("allproxies")
	dbProxies := database.GetAllProxies()
	// Show last time result when launch
	if proxies == nil && dbProxies != nil {
		cache.SetProxies("proxies", dbProxies)
		cache.LastCrawlTime = "抓取中，已载入上次数据库数据"
		log.Infoln("Database: loaded")
	}
	if dbProxies != nil {
		proxies = dbProxies.UniqAppendProxyList(proxies)
	}
	if proxies == nil {
		proxies = make(proxy.ProxyList, 0)
	}

	go func() {
		wg.Wait()
		close(pc)
	}() // Note: 为何并发？可以一边抓取一边读取而非抓完再读
	// for 用于阻塞goroutine
	for p := range pc { // Note: pc关闭后不能发送数据可以读取剩余数据
		if p != nil {
			proxies = proxies.UniqAppendProxy(p)
		}
	}

	proxies.NameClear()
	proxies = proxies.Derive()
	log.Infoln("CrawlGo unique proxy count: %d", len(proxies))

	// Clean Clash unsupported proxy because health check depends on clash
	proxies = provider.Clash{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Infoln("CrawlGo clash supported proxy count: %d", len(proxies))

	cache.SetProxies("allproxies", proxies)
	cache.AllProxiesCount = proxies.Len()
	log.Infoln("AllProxiesCount: %d", cache.AllProxiesCount)
	cache.SSProxiesCount = proxies.TypeLen("ss")
	log.Infoln("SSProxiesCount: %d", cache.SSProxiesCount)
	cache.SSRProxiesCount = proxies.TypeLen("ssr")
	log.Infoln("SSRProxiesCount: %d", cache.SSRProxiesCount)
	cache.VmessProxiesCount = proxies.TypeLen("vmess")
	log.Infoln("VmessProxiesCount: %d", cache.VmessProxiesCount)
	cache.TrojanProxiesCount = proxies.TypeLen("trojan")
	log.Infoln("TrojanProxiesCount: %d", cache.TrojanProxiesCount)
	cache.LastCrawlTime = time.Now().In(location).Format("2006-01-02 15:04:05")

	// Health Check
	log.Infoln("Now proceed proxy health check...")
	healthcheck.SpeedConn = C.Config.SpeedConnection
	healthcheck.DelayConn = C.Config.HealthCheckConnection
	if C.Config.HealthCheckTimeout > 0 {
		healthcheck.DelayTimeout = time.Second * time.Duration(C.Config.HealthCheckTimeout)
		log.Infoln("CONF: Health check timeout is set to %d seconds", C.Config.HealthCheckTimeout)
	}

	proxies = healthcheck.CleanBadProxiesWithGrpool(proxies)

	log.Infoln("CrawlGo clash usable proxy count: %d", len(proxies))

	// Format name like US_01 sorted by country
	proxies.NameAddCounrty().Sort()
	log.Infoln("Proxy rename DONE!")

	// Relay check and rename
	healthcheck.RelayCheck(proxies)
	for i := range proxies {
		if s, ok := healthcheck.ProxyStats.Find(proxies[i]); ok {
			if s.Relay {
				_, c, e := geoIp.GeoIpDB.Find(s.OutIp)
				if e == nil {
					proxies[i].SetName(fmt.Sprintf("Relay_%s-%s", proxies[i].BaseInfo().Name, c))
				}
			} else if s.Pool {
				proxies[i].SetName(fmt.Sprintf("Pool_%s", proxies[i].BaseInfo().Name))
			}
		}
	}

	proxies.NameAddIndex()

	// 可用节点存储
	cache.SetProxies("proxies", proxies)
	cache.UsefullProxiesCount = proxies.Len()
	database.SaveProxyList(proxies)
	database.ClearOldItems()

	log.Infoln("Usablility checking done. Open %s to check", C.Config.Domain+":"+C.Config.Port)

	// 测速
	speedTestNew(proxies)
	cache.SetString("clashproxies", provider.Clash{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.Provide()) // update static string provider
	cache.SetString("surgeproxies", provider.Surge{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
}

// Speed test for new proxies
func speedTestNew(proxies proxy.ProxyList) {
	if C.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if C.Config.SpeedTimeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(C.Config.SpeedTimeout)
			log.Infoln("config: Speed test timeout is set to %d seconds", C.Config.SpeedTimeout)
		}
		healthcheck.SpeedTestNew(proxies)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}

// Speed test for all proxies in proxy.ProxyList
func SpeedTest(proxies proxy.ProxyList) {
	if C.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if C.Config.SpeedTimeout > 0 {
			log.Infoln("config: Speed test timeout is set to %d seconds", C.Config.SpeedTimeout)
			healthcheck.SpeedTimeout = time.Second * time.Duration(C.Config.SpeedTimeout)
		}
		healthcheck.SpeedTestAll(proxies)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}
