package getter

import (
	"io/ioutil"
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
	body, err := ioutil.ReadAll(resp.Body)
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

//
// clash 文檔有效性檢查
//
func buildClashDoc(fullcheck bool, body []byte) []byte {
	// 這個正則表達式的設計很粗糙，對嵌套大括號 {{}} 檢測失靈，
	// 但這裏不用精確判斷，糊弄過去。
	regexp, _ := regexp.Compile(`{\s*name:[^,]+,\s*server:\s*\S+,\s*port:\s*\d+\s*,\s*type:[^}]+}`)

	tmp := strings.Split(strings.ReplaceAll(string(body), "\r\n", "\n"), "\n")
	var arr []string
	for index, s0 := range tmp {
		if !fullcheck && index == 0 && s0 == "proxies:" {
			// 如果第一行就是 "proxies:" 字符串, 就假設 body 是合法的 clash 訂閱文本,
			// 出於性能考慮, 不再進一步檢查, 直接返回.
			return body
		}
		match := regexp.FindStringIndex(s0)
		if match != nil {
			arr = append(arr, s0)
		}
	}

	if len(arr) == 0 {
		return []byte("")
	}
	arr = append([]string{"proxies:"}, arr...)

	return []byte(strings.Join(arr, "\n"))
}

func (c *Clash) Get2Chan(pc chan proxy.Proxy) {
	nodes := c.Get()
	log.Infoln("STATISTIC: Clash\tcount=%d\turl=%s", len(nodes), c.Url)
	for _, node := range nodes {
		pc <- node
	}
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
