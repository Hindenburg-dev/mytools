package proxy

import (
	"fmt"
	"net/http"
	"net/url"
)

// Proxy 根据 open 状态返回对应的 http.Client
func Proxyclient(open bool, proxyattr ...string) *http.Client {
	// 不开代理直接返回默认 Client，避免 else 嵌套
	if !open {
		return &http.Client{}
	}

	//默认地址：v2ray的监听端口
	URL := "http://127.0.0.1:10808"
	if len(proxyattr)>0&&proxyattr[0]!=""{
		URL = proxyattr[0]
	}

	proxyURL, err := url.Parse(URL)
	if err != nil {
		fmt.Println("代理地址解析失败:", err)
		return &http.Client{} // 发生解析错误时给一个保底的默认 Client，防止调用方拿到 nil 报空指针异常
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
}