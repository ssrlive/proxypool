package proxy

import (
	"fmt"
	"sort"
	"strings"
)

type ProxyList []Proxy

// sort排序使用
func (ps ProxyList) Len() int {
	return len(ps)
}

func (ps ProxyList) TypeLen(t string) int {
	l := 0
	for _, p := range ps {
		if p.TypeName() == t {
			l++
		}
	}
	return l
}

var sortType = make(map[string]int)

func init() {
	sortType["ss"] = 1
	sortType["ssr"] = 2
	sortType["vmess"] = 3
	sortType["trojan"] = 4
}

// sort排序使用
func (ps ProxyList) Less(i, j int) bool {
	if ps[i].BaseInfo().Country == ps[j].BaseInfo().Country {
		return sortType[ps[i].BaseInfo().Type] < sortType[ps[j].BaseInfo().Type]
	} else {
		return ps[i].BaseInfo().Country < ps[j].BaseInfo().Country
	}
}

// sort排序使用
func (ps ProxyList) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

// Deduplication by proxy identifier
func (ps ProxyList) Deduplication() ProxyList {
	result := make(ProxyList, 0, len(ps))
	temp := map[string]struct{}{}
	for _, item := range ps {
		if item != nil {
			if _, ok := temp[item.Identifier()]; !ok {
				temp[item.Identifier()] = struct{}{}
				result = append(result, item)
			}
		}
	}
	return result
}

func (ps ProxyList) Sort() ProxyList {
	sort.Sort(ps)
	return ps
}

func (ps ProxyList) NameSetCounrty() ProxyList {
	num := len(ps)
	for i := 0; i < num; i++ {
		ps[i].SetName(ps[i].BaseInfo().Country)
	}
	return ps
}

func (ps ProxyList) NameAddIndex() ProxyList {
	num := len(ps)
	for i := 0; i < num; i++ {
		ps[i].SetName(fmt.Sprintf("%s_%+02v", ps[i].BaseInfo().Name, i+1))
	}
	return ps
}

func (ps ProxyList) NameReIndex() ProxyList {
	num := len(ps)
	for i := 0; i < num; i++ {
		originName := ps[i].BaseInfo().Name
		country := strings.SplitN(originName, "_", 2)[0]
		ps[i].SetName(fmt.Sprintf("%s_%+02v", country, i+1))
	}
	return ps
}

func (ps ProxyList) NameAddTG() ProxyList {
	num := len(ps)
	for i := 0; i < num; i++ {
		ps[i].SetName(fmt.Sprintf("%s %s", ps[i].BaseInfo().Name, "TG@peekfun"))
	}
	return ps
}

func (ps ProxyList) Clone() ProxyList {
	result := make(ProxyList, 0, len(ps))
	for _, pp := range ps {
		if pp != nil {
			result = append(result, pp.Clone())
		}
	}
	return result
}

// Derive 将原有节点中的ss和ssr互相转换进行衍生
func (ps ProxyList) Derive() ProxyList {
	proxies := ps
	for _, p := range ps {
		if p == nil {
			continue
		}
		if p.TypeName() == "ss" {
			ssr, err := Convert2SSR(p)
			if err == nil {
				proxies = append(proxies, ssr)
			}
		} else if p.TypeName() == "ssr" {
			ss, err := Convert2SS(p)
			if err == nil {
				proxies = append(proxies, ss)
			}
		}
	}
	return proxies.Deduplication()
}

// Append unique new proxies to original ProxyList
func (ps *ProxyList) UniqAppendProxyList(new ProxyList) ProxyList {
	if len(new) == 0 {
		return *ps
	}
	if len(*ps) == 0 {
		return new
	}
	for _, p := range new {
		isExist := false
		for i, _ := range *ps {
			if (*ps)[i].Identifier() == p.Identifier() {
				isExist = true
				break
			}
		}
		if !isExist {
			*ps = append(*ps, p)
		}
	}
	return *ps
}

// Append an unique new proxy to original ProxyList
func (ps *ProxyList) UniqAppendProxy(new Proxy) ProxyList {
	if len(*ps) == 0 {
		*ps = append(*ps, new)
		return *ps
	}
	for i, _ := range *ps {
		if (*ps)[i].Identifier() == new.Identifier() {
			return *ps
		}
	}
	*ps = append(*ps, new)
	return *ps
}
