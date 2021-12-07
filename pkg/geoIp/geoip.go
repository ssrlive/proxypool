package geoIp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	bingeoip "github.com/ssrlive/proxypool/internal/bindata/geoip"
	"github.com/oschwald/geoip2-golang"
)

var GeoIpDB GeoIP

func InitGeoIpDB() error {
	geodb := "assets/GeoLite2-City.mmdb"
	// 判断文件是否存在
	_, err := os.Stat(geodb)
	if err != nil && os.IsNotExist(err) {
		err = bingeoip.RestoreAsset("", "assets/flags.json")
		if err != nil {
			panic(err)
		}
		err = bingeoip.RestoreAsset("", "assets/GeoLite2-City.mmdb")
		if err != nil {
			log.Println("文件不存在，请自行下载 Geoip2 City库，并保存在", geodb)
			panic(err)
		}
		GeoIpDB = NewGeoIP("assets/GeoLite2-City.mmdb", "assets/flags.json")
	}
	GeoIpDB = NewGeoIP("assets/GeoLite2-City.mmdb", "assets/flags.json")
	return nil
}

// GeoIP2
type GeoIP struct {
	db       *geoip2.Reader
	emojiMap map[string]string
}

type CountryEmoji struct {
	Code  string `json:"code"`
	Emoji string `json:"emoji"`
}

// new geoip from db file
func NewGeoIP(geodb, flags string) (geoip GeoIP) {
	// 运行到这里时geodb只能为存在
	db, err := geoip2.Open(geodb)
	if err != nil {
		log.Fatal(err)
	}
	geoip.db = db

	_, err = os.Stat(flags)
	if err != nil && os.IsNotExist(err) {
		log.Println("flags 文件不存在，请自行下载 flags.json，并保存在", flags)
		os.Exit(1)
	} else {
		data, err := ioutil.ReadFile(flags)
		if err != nil {
			log.Fatal(err)
			return
		}
		var countryEmojiList = make([]CountryEmoji, 0)
		err = json.Unmarshal(data, &countryEmojiList)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		emojiMap := make(map[string]string)
		for _, i := range countryEmojiList {
			emojiMap[i.Code] = i.Emoji
		}
		geoip.emojiMap = emojiMap
	}
	return
}

// find ip info
func (g GeoIP) Find(ipORdomain string) (ip, country string, err error) {
	ips, err := net.LookupIP(ipORdomain)
	if err != nil {
		return "", "", err
	}
	ip = ips[0].String()

	var record *geoip2.City
	record, err = g.db.City(ips[0])
	if err != nil {
		return
	}
	countryIsoCode := record.Country.IsoCode
	emoji, found := g.emojiMap[countryIsoCode]
	if found {
		country = fmt.Sprintf("%v%v", emoji, countryIsoCode)
	} else {
		country = "��ZZ"
	}
	return ip, country, err
}
