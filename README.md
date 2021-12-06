<h1 align="center">
  <br>proxypool<br>
</h1>

<h5 align="center">自动抓取tg频道、订阅地址、公开互联网上的ss、ssr、vmess、trojan节点信息，聚合去重测试可用性后提供节点列表</h5>

<p align="center">
  <a href="https://github.com/ssrlive/proxypool/actions">
    <img src="https://img.shields.io/github/workflow/status/zu1k/proxypool/Go?style=flat-square" alt="Github Actions">
  </a>
  <a href="https://goreportcard.com/report/github.com/zu1k/proxypool">
    <img src="https://goreportcard.com/badge/github.com/zu1k/proxypool?style=flat-square">
  </a>
  <a href="https://github.com/ssrlive/proxypool/releases">
    <img src="https://img.shields.io/github/release/zu1k/proxypool/all.svg?style=flat-square">
  </a>
</p>

## 支持

- 支持ss、ssr、vmess、trojan多种类型
- Telegram频道抓取
- 订阅地址抓取解析
- 公开互联网页面模糊抓取
- 定时抓取自动更新
- 通过配置文件设置抓取源
- 自动检测节点可用性
- 提供clash、surge配置文件
- 提供ss、ssr、vmess、sip002订阅

## 安装

以下四选一。

### 使用Heroku

点击按钮进入部署页面，填写基本信息然后运行

其中 `DOMAIN` 需要填写为你需要绑定的域名，`CONFIG_FILE` 需要填写你的配置文件路径。

> heroku app域名为appname.herokuapp.com。项目内配置文件为./config/config.yaml

配置文件模板见 config/config.yaml 文件，可选项区域均可不填。完整配置选项请查看[配置文件说明](https://github.com/Sansui233/proxypool/wiki/%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E8%AF%B4%E6%98%8E)。

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy)

> 因为爬虫程序需要持续运行，所以至少选择 $7/月 的配置
> 免费配置长时间无人访问会被heroku强制停止

### 从源码编译

需要 [安装 Golang](https://golang.org/doc/install) ， 然后拉取代码,

```bash
go get -u -v github.com/ssrlive/proxypool@latest
```

或者，拉取代码的另一种方式 
```
git clone https://github.com/ssrlive/proxypool.git
cd proxypool
go get
go build
```
then edit `config/config.yaml` and `config/source.yaml` and run it
```
./proxypool -c ./config/config.yaml
```
或者从源代码运行
```
go run main.go -c ./config/config.yaml
```


### 下载预编译程序

从这里下载预编译好的程序 [release](https://github.com/ssrlive/proxypool/releases)

### 使用docker

```sh
docker pull docker.pkg.github.com/ssrlive/proxypool/proxypool:latest
```

## 使用

运行该程序需要具有访问完整互联网的能力。

### 修改配置文件

首先修改 config.yaml 中的必要配置信息，cf开头的选项不需要填写

source.yaml 文件中定义了抓取源，需要定期手动维护更新

完整的配置选项见[配置文件说明](https://github.com/Sansui233/proxypool/wiki/%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6%E8%AF%B4%E6%98%8E)

### 启动程序

使用 `-c` 参数指定配置文件路径，支持http链接

```shell
proxypool -c ./config/config.yaml
```

如果需要部署到VPS，更多细节请[查看wiki](https://github.com/Sansui233/proxypool/wiki/%E9%83%A8%E7%BD%B2%E5%88%B0VPS-Step-by-Step)。

## Clash配置文件

远程部署时Clash配置文件访问：https://domain/clash/config

本地运行时Clash配置文件访问：http://127.0.0.1:[端口]/clash/localconfig

## 本地检查节点可用性

此项非必须。为了提高实际可用性，可选择增加一个本地服务器，检测远程proxypool节点在本地的可用性并提供配置，见[proxypoolCheck](https://github.com/Sansui233/proxypoolCheck)。

## 截图

![Speedtest](docs/speedtest.png)

![Fast](docs/fast.png)

## 声明

本项目遵循 GNU General Public License v3.0 开源，在此基础上，所有使用本项目提供服务者都必须在网站首页保留指向本项目的链接

本项目仅限个人自己使用，禁止使用本项目进行营利和做其他违法事情，产生的一切后果本项目概不负责
