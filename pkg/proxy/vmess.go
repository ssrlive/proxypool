package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ssrlive/proxypool/pkg/tool"
)

var (
	ErrorNotVmessLink          = errors.New("not a correct vmess link")
	ErrorVmessPayloadParseFail = errors.New("vmess link payload parse failed")
)

type Vmess struct {
	Base
	UUID           string       `yaml:"uuid" json:"uuid"`
	AlterID        int          `yaml:"alterId" json:"alterId"`
	Cipher         string       `yaml:"cipher" json:"cipher"`
	Network        string       `yaml:"network,omitempty" json:"network,omitempty"`
	ServerName     string       `yaml:"servername,omitempty" json:"servername,omitempty"`
	HTTPOpts       HTTPOptions  `yaml:"http-opts,omitempty" json:"http-opts,omitempty"`
	HTTP2Opts      HTTP2Options `yaml:"h2-opts,omitempty" json:"h2-opts,omitempty"`
	TLS            bool         `yaml:"tls,omitempty" json:"tls,omitempty"`
	SkipCertVerify bool         `yaml:"skip-cert-verify,omitempty" json:"skip-cert-verify,omitempty"`
	WSOpts         *WSOptions   `yaml:"ws-opts,omitempty" json:"ws-opts,omitempty"`
}

type WSOptions struct {
	Path    string            `yaml:"path,omitempty" json:"path,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

type HTTPOptions struct {
	Method  string              `yaml:"method,omitempty" json:"method,omitempty"`
	Path    []string            `yaml:"path,omitempty" json:"path,omitempty"`
	Headers map[string][]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

type HTTP2Options struct {
	Host []string `yaml:"host,omitempty" json:"host,omitempty"`
	Path string   `yaml:"path,omitempty" json:"path,omitempty"` // 暂只处理一个Path
}

func (v *Vmess) UnmarshalJSON(data []byte) error {
	tmp := struct {
		Base
		UUID           string            `yaml:"uuid" json:"uuid"`
		AlterID        int               `yaml:"alterId" json:"alterId"`
		Cipher         string            `yaml:"cipher" json:"cipher"`
		Network        string            `yaml:"network,omitempty" json:"network,omitempty"`
		ServerName     string            `yaml:"servername,omitempty" json:"servername,omitempty"`
		HTTPOpts       HTTPOptions       `yaml:"http-opts,omitempty" json:"http-opts,omitempty"`
		HTTP2Opts      HTTP2Options      `yaml:"h2-opts,omitempty" json:"h2-opts,omitempty"`
		TLS            bool              `yaml:"tls,omitempty" json:"tls,omitempty"`
		SkipCertVerify bool              `yaml:"skip-cert-verify,omitempty" json:"skip-cert-verify,omitempty"`
		WSOpts         WSOptions         `yaml:"ws-opts,omitempty" json:"ws-opts,omitempty"`
		WSPath         string            `yaml:"ws-path,omitempty" json:"ws-path,omitempty"`
		WSHeaders      map[string]string `yaml:"ws-headers,omitempty" json:"ws-headers,omitempty"`
	}{}

	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	v.Base = tmp.Base
	v.UUID = tmp.UUID
	v.AlterID = tmp.AlterID
	v.Cipher = tmp.Cipher
	v.Network = tmp.Network
	v.ServerName = tmp.ServerName
	v.HTTPOpts = tmp.HTTPOpts
	v.HTTP2Opts = tmp.HTTP2Opts
	v.TLS = tmp.TLS
	v.SkipCertVerify = tmp.SkipCertVerify
	if tmp.Network == "ws" {
		if tmp.WSOpts.Path == "" {
			tmp.WSOpts.Path = tmp.WSPath
		}
		if tmp.WSOpts.Headers == nil {
			tmp.WSOpts.Headers = tmp.WSHeaders
		}
		v.WSOpts = &WSOptions{
			Path:    tmp.WSOpts.Path,
			Headers: tmp.WSOpts.Headers,
		}
	}

	return nil
}

func (v Vmess) Identifier() string {
	return net.JoinHostPort(v.Server, strconv.Itoa(v.Port)) + v.Cipher + v.UUID
}

func (v Vmess) String() string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func (v Vmess) ToClash() string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return "- " + string(data)
}

func (v Vmess) ToSurge() string {
	// node2 = vmess, server, port, username=, ws=true, ws-path=, ws-headers=
	if v.Network == "ws" {
		wsHeasers := ""
		for k, v := range v.WSOpts.Headers {
			if wsHeasers == "" {
				wsHeasers = k + ":" + v
			} else {
				wsHeasers += "|" + k + ":" + v
			}
		}
		text := fmt.Sprintf("%s = vmess, %s, %d, username=%s, ws=true, tls=%t, ws-path=%s",
			v.Name, v.Server, v.Port, v.UUID, v.TLS, v.WSOpts.Path)
		if wsHeasers != "" {
			text += ", ws-headers=" + wsHeasers
		}
		return text
	} else {
		return fmt.Sprintf("%s = vmess, %s, %d, username=%s, tls=%t",
			v.Name, v.Server, v.Port, v.UUID, v.TLS)
	}
}

func (v Vmess) Clone() Proxy {
	return &v
}

func (v Vmess) Link() (link string) {
	vjv, err := json.Marshal(v.toLinkJson())
	if err != nil {
		return
	}
	return fmt.Sprintf("vmess://%s", tool.Base64EncodeBytes(vjv))
}

type vmessLinkJson struct {
	Add  string `json:"add"`
	V    string `json:"v"`
	Ps   string `json:"ps"`
	Port int    `json:"port"`
	Id   string `json:"id"`
	Aid  string `json:"aid"`
	Net  string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	Tls  string `json:"tls"`
}

func (v Vmess) toLinkJson() vmessLinkJson {
	vj := vmessLinkJson{
		Add:  v.Server,
		Ps:   v.Name,
		Port: v.Port,
		Id:   v.UUID,
		Aid:  strconv.Itoa(v.AlterID),
		Net:  v.Network,
		Path: v.WSOpts.Path,
		Host: v.ServerName,
		V:    "2",
	}
	if v.TLS {
		vj.Tls = "tls"
	}
	if host, ok := v.WSOpts.Headers["HOST"]; ok && host != "" {
		vj.Host = host
	}
	return vj
}

func ParseVmessLink(link string) (*Vmess, error) {
	if !strings.HasPrefix(link, "vmess") {
		return nil, ErrorNotVmessLink
	}

	vmessmix := strings.SplitN(link, "://", 2)
	if len(vmessmix) < 2 {
		return nil, ErrorNotVmessLink
	}
	linkPayload := vmessmix[1]
	if strings.Contains(linkPayload, "?") {
		// 使用第二种解析方法 目测是Shadowrocket格式
		var infoPayloads []string
		if strings.Contains(linkPayload, "/?") {
			infoPayloads = strings.SplitN(linkPayload, "/?", 2)
		} else {
			infoPayloads = strings.SplitN(linkPayload, "?", 2)
		}
		if len(infoPayloads) < 2 {
			return nil, ErrorNotVmessLink
		}

		baseInfo, err := tool.Base64DecodeString(infoPayloads[0])
		if err != nil {
			return nil, ErrorVmessPayloadParseFail
		}
		baseInfoPath := strings.Split(baseInfo, ":")
		if len(baseInfoPath) < 3 {
			return nil, ErrorPathNotComplete
		}
		// base info
		cipher := baseInfoPath[0]
		mixInfo := strings.SplitN(baseInfoPath[1], "@", 2)
		if len(mixInfo) < 2 {
			return nil, ErrorVmessPayloadParseFail
		}
		uuid := mixInfo[0]
		server := mixInfo[1]
		portStr := baseInfoPath[2]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, ErrorVmessPayloadParseFail
		}

		moreInfo, _ := url.ParseQuery(infoPayloads[1])
		remarks := moreInfo.Get("remarks")

		// Transmission protocol
		wsHeaders := make(map[string]string)
		h2Opt := HTTP2Options{
			Host: make([]string, 0),
		}
		httpOpt := HTTPOptions{}

		// Network <- obfs=websocket
		obfs := moreInfo.Get("obfs")
		network := "tcp"
		if obfs == "http" {
			httpOpt.Method = "GET" // 不知道Headers为空时会不会报错
		}
		if obfs == "websocket" {
			network = "ws"
		} else { // when http h2
			network = obfs
		}
		// HTTP Object: Host <- obfsParam=www.036452916.xyz
		host := moreInfo.Get("obfsParam")
		if host != "" {
			switch obfs {
			case "websocket":
				wsHeaders["Host"] = host
			case "h2":
				h2Opt.Host = append(h2Opt.Host, host)
			}
		}
		// HTTP Object: Path
		path := moreInfo.Get("path")
		if path == "" {
			path = "/"
		}
		switch obfs {
		case "h2":
			h2Opt.Path = path
			path = ""
		case "http":
			httpOpt.Path = append(httpOpt.Path, path)
			path = ""
		}

		tls := moreInfo.Get("tls") == "1"
		if obfs == "h2" {
			tls = true
		}
		// allowInsecure=1 Clash config unsuported
		// alterId=64
		aid := 0
		aidStr := moreInfo.Get("alterId")
		if aidStr != "" {
			aid, _ = strconv.Atoi(aidStr)
		}

		v := Vmess{
			Base: Base{
				Name:   remarks + "_" + strconv.Itoa(rand.Int()),
				Server: server,
				Port:   port,
				Type:   "vmess",
				UDP:    false,
			},
			UUID:           uuid,
			AlterID:        aid,
			Cipher:         cipher,
			TLS:            tls,
			Network:        network,
			HTTPOpts:       httpOpt,
			HTTP2Opts:      h2Opt,
			SkipCertVerify: true,
			ServerName:     server,
		}
		if v.Network == "ws" {
			v.WSOpts = &WSOptions{
				Path:    path,
				Headers: wsHeaders,
			}
		}

		return &v, nil
	} else {
		// V2rayN ref: https://github.com/2dust/v2rayN/wiki/%E5%88%86%E4%BA%AB%E9%93%BE%E6%8E%A5%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E(ver-2)
		payload, err := tool.Base64DecodeString(linkPayload)
		if err != nil {
			return nil, ErrorVmessPayloadParseFail
		}
		vmessJson := vmessLinkJson{}
		jsonMap, err := str2jsonDynaUnmarshal(payload)
		if err != nil {
			return nil, err
		}
		vmessJson, err = mapStrInter2VmessLinkJson(jsonMap)
		if err != nil {
			return nil, err
		}

		alterId, err := strconv.Atoi(vmessJson.Aid)
		if err != nil {
			alterId = 0
		}
		tls := vmessJson.Tls == "tls"

		if vmessJson.Net == "h2" {
			tls = true
		}

		wsHeaders := make(map[string]string)
		h2Opt := HTTP2Options{}
		httpOpt := HTTPOptions{}

		if vmessJson.Net == "http" {
			httpOpt.Method = "GET" // 不知道Headers为空时会不会报错
		}

		if vmessJson.Host != "" {
			switch vmessJson.Net {
			case "h2":
				h2Opt.Host = append(h2Opt.Host, vmessJson.Host) // 不知道为空时会不会报错
			case "ws":
				wsHeaders["HOST"] = vmessJson.Host
			}
		}

		if vmessJson.Path == "" {
			vmessJson.Path = "/"
		}
		switch vmessJson.Net {
		case "h2":
			h2Opt.Path = vmessJson.Path
			vmessJson.Path = ""
		case "http":
			httpOpt.Path = append(httpOpt.Path, vmessJson.Path)
			vmessJson.Path = ""
		}

		v := Vmess{
			Base: Base{
				Name:   "",
				Server: vmessJson.Add,
				Port:   vmessJson.Port,
				Type:   "vmess",
				UDP:    false,
			},
			UUID:           vmessJson.Id,
			AlterID:        alterId,
			Cipher:         "auto",
			Network:        vmessJson.Net,
			HTTPOpts:       httpOpt,
			HTTP2Opts:      h2Opt,
			ServerName:     vmessJson.Host,
			TLS:            tls,
			SkipCertVerify: true,
		}

		if v.Network == "ws" {
			v.WSOpts = &WSOptions{
				Path:    vmessJson.Path,
				Headers: wsHeaders,
			}
		}

		return &v, nil
	}
}

var (
	vmessPlainRe = regexp.MustCompile("vmess://([A-Za-z0-9+/_?&=-])+")
)

func GrepVmessLinkFromString(text string) []string {
	results := make([]string, 0)
	if !strings.Contains(text, "vmess://") {
		return results
	}
	texts := strings.Split(text, "vmess://")
	for _, text := range texts {
		results = append(results, vmessPlainRe.FindAllString("vmess://"+text, -1)...)
	}
	return results
}

func str2jsonDynaUnmarshal(s string) (jsn map[string]interface{}, err error) {
	var f interface{}
	err = json.Unmarshal([]byte(s), &f)
	if err != nil {
		return nil, err
	}
	jsn, ok := f.(map[string]interface{}) // f is pointer point to map struct
	if !ok {
		return nil, ErrorVmessPayloadParseFail
	}
	return jsn, err
}

func mapStrInter2VmessLinkJson(jsn map[string]interface{}) (vmessLinkJson, error) {
	vmess := vmessLinkJson{}
	var err error

	vmessVal := reflect.ValueOf(&vmess).Elem()
	for i := 0; i < vmessVal.NumField(); i++ {
		tags := vmessVal.Type().Field(i).Tag.Get("json")
		tag := strings.Split(tags, ",")
		if jsnVal, ok := jsn[strings.ToLower(tag[0])]; ok {
			if strings.ToLower(tag[0]) == "port" { // set int in port
				switch jsnVal := jsnVal.(type) {
				case float64:
					vmessVal.Field(i).SetInt(int64(jsnVal))
				case string: // Force Convert
					valInt, err := strconv.Atoi(jsnVal)
					if err != nil {
						valInt = 443
					}
					vmessVal.Field(i).SetInt(int64(valInt))
				default:
					vmessVal.Field(i).SetInt(443)
				}
			} else if strings.ToLower(tag[0]) == "ps" {
				continue
			} else { // set string in other fields
				switch jsnVal := jsnVal.(type) {
				case string:
					vmessVal.Field(i).SetString(jsnVal)
				default: // Force Convert
					vmessVal.Field(i).SetString(fmt.Sprintf("%v", jsnVal))
				}
			}
		}
	}
	return vmess, err
}
