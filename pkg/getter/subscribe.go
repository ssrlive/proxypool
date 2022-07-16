package getter

import (
	"io/ioutil"
	"strings"
	"sync"

	"github.com/ssrlive/proxypool/log"

	"github.com/ssrlive/proxypool/pkg/proxy"
	"github.com/ssrlive/proxypool/pkg/tool"
)

// Add key value pair to creatorMap(string → creator) in base.go
func init() {
	Register("subscribe", NewSubscribe)
}

// Subscribe is A Getter with an additional property
type Subscribe struct {
	Url string
}

// Get() of Subscribe is to implement Getter interface
func (s *Subscribe) Get() proxy.ProxyList {
	resp, err := tool.GetHttpClient().Get(s.Url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	nodesString, err := tool.Base64DecodeString(string(body))
	if err != nil {
		return nil
	}
	nodesString = strings.ReplaceAll(nodesString, "\t", "")

	nodes := strings.Split(nodesString, "\n")
	return StringArray2ProxyArray(nodes)
}

// Subscribe is to implement Getter interface. It gets proxies and send proxy to channel one by one
func (s *Subscribe) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := s.Get()
	log.Infoln("STATISTIC: Subscribe\tcount=%d\turl=%s", len(nodes), s.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func NewSubscribe(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &Subscribe{
			Url: url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}
