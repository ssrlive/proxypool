package api

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/asdlokj1qpi23/proxypool/log"
	"golang.org/x/exp/slices"

	"github.com/asdlokj1qpi23/proxypool/config"
	appcache "github.com/asdlokj1qpi23/proxypool/internal/cache"
	"github.com/asdlokj1qpi23/proxypool/pkg/provider"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

const version = "v0.7.15"

var router *gin.Engine

func setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	router = gin.New() // 没有任何中间件的路由
	store := persistence.NewInMemoryStore(time.Minute)
	router.Use(gin.Recovery(), cache.SiteCache(store, time.Minute)) // 加上处理panic的中间件，防止遇到panic退出程序

	temp, err := loadHTMLTemplate() // 加载html模板，模板源存放于html.go中的类似_assetsHtmlSurgeHtml的变量
	if err != nil {
		panic(err)
	}
	router.SetHTMLTemplate(temp) // 应用模板

	router.StaticFile("/static/index.js", "assets/static/index.js")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/index.html", gin.H{
			"domain":               config.Config.Domain,
			"getters_count":        appcache.GettersCount,
			"all_proxies_count":    appcache.AllProxiesCount,
			"ss_proxies_count":     appcache.SSProxiesCount,
			"ssr_proxies_count":    appcache.SSRProxiesCount,
			"vmess_proxies_count":  appcache.VmessProxiesCount,
			"trojan_proxies_count": appcache.TrojanProxiesCount,
			"useful_proxies_count": appcache.UsefullProxiesCount,
			"last_crawl_time":      appcache.LastCrawlTime,
			"is_speed_test":        appcache.IsSpeedTest,
			"is_netflix_test":      appcache.IsNetflixTest,
			"is_disney_test":       appcache.IsDisneyTest,
			"netflix_count":        appcache.NetflixCount,
			"disney_count":         appcache.DisneyCount,
			"version":              version,
		})
	})

	router.GET("/clash", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash.html", gin.H{
			"domain": config.Config.Domain,
			"port":   config.Config.Port,
		})
	})

	router.GET("/surge", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/shadowrocket", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/shadowrocket.html", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash-config.yaml", gin.H{
			"domain": config.Config.Domain,
		})
	})
	router.GET("/clash/localconfig", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/clash-config-local.yaml", gin.H{
			"port": config.Config.Port,
		})
	})

	router.GET("/surge/config", func(c *gin.Context) {
		c.HTML(http.StatusOK, "assets/html/surge.conf", gin.H{
			"domain": config.Config.Domain,
		})
	})

	router.GET("/clash/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		streamFilter := c.DefaultQuery("stream", "")
		streamNotFilter := c.DefaultQuery("nstream", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" && proxySpeed == "" && proxyFilter == "" && streamFilter == "" && streamNotFilter == "" {
			text = appcache.GetString("clashproxies") // A string. To show speed in this if condition, this must be updated after speedtest
			if text == "" {
				proxies := appcache.GetProxies("proxies")
				clash := provider.Clash{
					Base: provider.Base{
						Proxies: &proxies,
					},
				}
				text = clash.Provide() // 根据Query筛选节点
				appcache.SetString("clashproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := appcache.GetProxies("allproxies")
			clash := provider.Clash{
				Base: provider.Base{
					Proxies:         &proxies,
					Types:           proxyTypes,
					Country:         proxyCountry,
					NotCountry:      proxyNotCountry,
					Speed:           proxySpeed,
					Filter:          proxyFilter,
					StreamFilter:    streamFilter,
					StreamNotFilter: streamNotFilter,
				},
			}
			text = clash.Provide() // 根据Query筛选节点
		} else {
			proxies := appcache.GetProxies("proxies")
			clash := provider.Clash{
				Base: provider.Base{
					Proxies:         &proxies,
					Types:           proxyTypes,
					Country:         proxyCountry,
					NotCountry:      proxyNotCountry,
					Speed:           proxySpeed,
					Filter:          proxyFilter,
					StreamFilter:    streamFilter,
					StreamNotFilter: streamNotFilter,
				},
			}
			text = clash.Provide() // 根据Query筛选节点
		}
		c.String(200, text)
	})
	router.GET("/surge/proxies", func(c *gin.Context) {
		proxyTypes := c.DefaultQuery("type", "")
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		text := ""
		if proxyTypes == "" && proxyCountry == "" && proxyNotCountry == "" && proxySpeed == "" {
			text = appcache.GetString("surgeproxies") // A string. To show speed in this if condition, this must be updated after speedtest
			if text == "" {
				proxies := appcache.GetProxies("proxies")
				surge := provider.Surge{
					Base: provider.Base{
						Proxies: &proxies,
					},
				}
				text = surge.Provide()
				appcache.SetString("surgeproxies", text)
			}
		} else if proxyTypes == "all" {
			proxies := appcache.GetProxies("allproxies")
			surge := provider.Surge{
				Base: provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
					Speed:      proxySpeed,
					Filter:     proxyFilter,
				},
			}
			text = surge.Provide()
		} else {
			proxies := appcache.GetProxies("proxies")
			surge := provider.Surge{
				Base: provider.Base{
					Proxies:    &proxies,
					Types:      proxyTypes,
					Country:    proxyCountry,
					NotCountry: proxyNotCountry,
					Filter:     proxyFilter,
				},
			}
			text = surge.Provide()
		}
		c.String(200, text)
	})

	router.GET("/ss/sub", func(c *gin.Context) {
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		proxies := appcache.GetProxies("proxies")
		ssSub := provider.SSSub{
			Base: provider.Base{
				Proxies:    &proxies,
				Types:      "ss",
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
				Speed:      proxySpeed,
				Filter:     proxyFilter,
			},
		}
		c.String(200, ssSub.Provide())
	})
	router.GET("/ssr/sub", func(c *gin.Context) {
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		proxies := appcache.GetProxies("proxies")
		ssrSub := provider.SSRSub{
			Base: provider.Base{
				Proxies:    &proxies,
				Types:      "ssr",
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
				Speed:      proxySpeed,
				Filter:     proxyFilter,
			},
		}
		c.String(200, ssrSub.Provide())
	})
	router.GET("/vmess/sub", func(c *gin.Context) {
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		proxies := appcache.GetProxies("proxies")
		vmessSub := provider.VmessSub{
			Base: provider.Base{
				Proxies:    &proxies,
				Types:      "vmess",
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
				Speed:      proxySpeed,
				Filter:     proxyFilter,
			},
		}
		c.String(200, vmessSub.Provide())
	})
	router.GET("/sip002/sub", func(c *gin.Context) {
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		proxies := appcache.GetProxies("proxies")
		sip002Sub := provider.SIP002Sub{
			Base: provider.Base{
				Proxies:    &proxies,
				Types:      "ss",
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
				Speed:      proxySpeed,
				Filter:     proxyFilter,
			},
		}
		c.String(200, sip002Sub.Provide())
	})
	router.GET("/trojan/sub", func(c *gin.Context) {
		proxyCountry := c.DefaultQuery("c", "")
		proxyNotCountry := c.DefaultQuery("nc", "")
		proxySpeed := c.DefaultQuery("speed", "")
		proxyFilter := c.DefaultQuery("filter", "")
		proxies := appcache.GetProxies("proxies")
		trojanSub := provider.TrojanSub{
			Base: provider.Base{
				Proxies:    &proxies,
				Types:      "trojan",
				Country:    proxyCountry,
				NotCountry: proxyNotCountry,
				Speed:      proxySpeed,
				Filter:     proxyFilter,
			},
		}
		c.String(200, trojanSub.Provide())
	})
	router.GET("/link/:id", func(c *gin.Context) {
		idx := c.Param("id")
		proxies := appcache.GetProxies("allproxies")
		id, err := strconv.Atoi(idx)
		if err != nil {
			c.String(500, err.Error())
		}
		if id >= proxies.Len() || id < 0 {
			c.String(500, "id out of range")
		}
		c.String(200, proxies[id].Link())
	})
}

func Run() {
	setupRouter()
	servePort := config.Config.Port
	envp := os.Getenv("PORT") // environment port for heroku app
	if envp != "" {
		servePort = envp
	}
	// Run on this server
	err := router.Run(":" + servePort)
	if err != nil {
		log.Errorln("router: Web server starting failed. Make sure your port %s has not been used. \n%s", servePort, err.Error())
	} else {
		log.Infoln("Proxypool is serving on port: %s", servePort)
	}
}

// 返回页面templates
func loadHTMLTemplate() (t *template.Template, err error) {
	t = template.New("")
	for _, fileName := range AssetNames() { //fileName带有路径前缀
		if strings.Contains(fileName, "css") {
			continue
		}
		data := MustAsset(fileName)                  //读取页面数据
		t, err = t.New(fileName).Parse(string(data)) //生成带路径名称的模板
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	var _bindata = []string{
		"assets/html/clash-config-local.yaml",
		"assets/html/clash-config.yaml",
		"assets/html/clash.html",
		"assets/html/index.html",
		"assets/html/shadowrocket.html",
		"assets/html/surge.conf",
		"assets/html/surge.html",
		"assets/static/index.js",
	}
	return _bindata
}

func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}
	return a
}

func Asset(name string) ([]byte, error) {
	var _bindata = AssetNames()
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if slices.Contains(_bindata, cannonicalName) {
		parentPath := config.ResourceRoot()
		fullFilePath := filepath.Join(parentPath, cannonicalName)
		contents, err := os.ReadFile(fullFilePath)
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return contents, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}
