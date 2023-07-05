package stream

import (
	"github.com/asdlokj1qpi23/proxypool/pkg/proxy"
	"strings"
)

func NETFLIXFilter(netflixList []Element, proxies proxy.ProxyList) (resultProxies proxy.ProxyList) {
	for _, v := range proxies {
		item := v
		for idx := range netflixList {
			var element = netflixList[idx]
			if element.Name == v.BaseInfo().Name {
				v.SetName("netflix_" + element.Country)
				item = v
				break
			}
		}
		resultProxies = append(resultProxies, item)
	}
	return
}

func DISNEYFilter(disneyList []Element, proxies proxy.ProxyList) (resultProxies proxy.ProxyList) {
	for _, v := range proxies {
		item := v
		for idx := range disneyList {
			var element = disneyList[idx]
			if element.Name == v.BaseInfo().Name {
				if strings.Contains(v.BaseInfo().Name, "netflix_") {
					v.SetName("disney_" + v.BaseInfo().Name)
				} else {
					v.SetName("disney_" + element.Country)
				}
				item = v
				break
			}
		}
		resultProxies = append(resultProxies, item)
	}
	return
}
