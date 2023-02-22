package getter

import (
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/ssrlive/proxypool/log"
	"github.com/ssrlive/proxypool/pkg/proxy"
	"github.com/ssrlive/proxypool/pkg/tool"
	"gopkg.in/yaml.v3"
)

func init() {
	Register("clash", NewClashGetter)
}

type Clash struct {
	Url string
}

type config struct {
	Proxy []map[string]interface{} `json:"proxies" yaml:"proxies"`
}

func (c *Clash) Get() proxy.ProxyList {
	resp, err := tool.GetHttpClient().Get(c.Url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	body = buildClashDoc(false, body)

	conf := config{}
	err = yaml.Unmarshal(body, &conf)
	if err != nil {
		return nil
	}

	return ClashProxy2ProxyArray(conf.Proxy)

}

// clash 文檔有效性檢查
func buildClashDoc(fullcheck bool, body []byte) []byte {
	regexp0, _ := regexp.Compile(`-\s*{`)

	tmp := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	var arr []string
	for index, s0 := range tmp {
		if !fullcheck && index == 0 && s0 == "proxies:" {
			// 如果第一行就是 "proxies:" 字符串, 就假設 body 是合法的 clash 訂閱文本,
			// 出於性能考慮, 不再進一步檢查, 直接返回.
			return body
		}
		if index == 0 {
			// 如果第一行是 "port: 7890" 字样，就认定为 clash 格式，直接返回。
			regexp2, _ := regexp.Compile(`^port:\s*[0-9]+`)
			match2 := regexp2.FindStringIndex(s0)
			if match2 != nil {
				return body
			}
		}

		match := regexp0.FindStringIndex(s0)
		if match == nil {
			continue
		}
		nodeStr := s0[match[1]-1:]
		pmap := make(map[string]interface{})
		if json.Unmarshal([]byte(nodeStr), &pmap) != nil {
			continue
		}

		arr = append(arr, "  - "+nodeStr)
	}

	if len(arr) == 0 {
		return []byte("")
	}
	arr = append([]string{"proxies:"}, arr...)

	return []byte(strings.Join(arr, "\n"))
}

func (c *Clash) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := c.Get()
	log.Infoln("STATISTIC: Clash\tcount=%d\turl=%s", len(nodes), c.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func NewClashGetter(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &Clash{
			Url: url,
		}, nil
	}
	return nil, ErrorUrlNotFound
}
