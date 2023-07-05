package stream

import (
	"fmt"
	C "github.com/Dreamacro/clash/constant"
	"net/url"
)

func urlToMetadata(rawURL string) (addr C.Metadata, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}

	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			err = fmt.Errorf("%s scheme not Support", rawURL)
			return
		}
	}
	addr = C.Metadata{
		Host:    u.Hostname(),
		DstIP:   nil,
		DstPort: port,
	}
	return
}
