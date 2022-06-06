package provider

import (
	"strings"

	"github.com/Sansui233/proxypool/pkg/tool"
)

type SSRSub struct {
	Base
}

func (sub SSRSub) Provide() string {
	sub.Types = "ssr"
	sub.preFilter()
	var resultBuilder strings.Builder
	for _, p := range *sub.Proxies {
		resultBuilder.WriteString(p.Link() + "\n")
	}
	return tool.Base64EncodeString(resultBuilder.String(), false)
}
