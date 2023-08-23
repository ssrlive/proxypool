package stream

import (
	"fmt"
	C "github.com/Dreamacro/clash/constant"
	"net/url"
	"strconv"
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
		DstPort: C.Port(convertPort(port)),
	}
	return
}

func convertPort(port string) (uint16num uint16) {
	num, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		fmt.Println("转换失败：", err)
		return
	}
	uint16num = uint16(num)
	return
}
