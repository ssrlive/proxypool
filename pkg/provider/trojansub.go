package provider

import (
	"strings"

	"github.com/asdlokj1qpi23/proxypool/pkg/tool"
)

type TrojanSub struct {
	Base
}

func (sub TrojanSub) Provide() string {
	sub.Types = "trojan"
	sub.preFilter()
	var resultBuilder strings.Builder
	for _, p := range *sub.Proxies {
		resultBuilder.WriteString(p.Link() + "\n")
	}
	return tool.Base64EncodeString(resultBuilder.String(), false)
}
