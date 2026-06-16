package utils

import (
	"errors"
	"net/url"
	"strings"
)

func IsAllowedDownloadURL(raw string) error {
	_, err := ResourcePlatform(raw)
	return err
}

func ResourcePlatform(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("只支持 http 或 https 地址")
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return "", errors.New("缺少资源域名")
	}

	if matchesHost(host, "xiaohongshu.com", "xhslink.com", "xhscdn.com") || (strings.HasPrefix(host, "sns-") && strings.HasSuffix(host, ".xhscdn.com")) {
		return "xhs", nil
	}
	if matchesHost(host, "douyin.com", "iesdouyin.com", "snssdk.com", "qichuangtianxia.com", "douyinvod.com", "douyinpic.com", "zjcdn.com") {
		return "douyin", nil
	}
	if matchesHost(host, "kuaishou.com", "gifshow.com", "chenzhongtech.com", "yximgs.com", "ksapisrv.com", "kwaicdn.com") {
		return "kuaishou", nil
	}
	if matchesHost(host, "miyoushe.com", "mihoyo.com") {
		return "miyoushe", nil
	}
	if matchesHost(host, "doubao.com", "byteimg.com", "douyinpic.com", "douyinvod.com") {
		return "doubao", nil
	}

	return "", errors.New("资源域名不在允许列表")
}

func matchesHost(host string, domains ...string) bool {
	for _, domain := range domains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func DetectPlatformByHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())

	if host == "kuaishou.com" || strings.HasSuffix(host, ".kuaishou.com") ||
		host == "gifshow.com" || strings.HasSuffix(host, ".gifshow.com") ||
		host == "chenzhongtech.com" || strings.HasSuffix(host, ".chenzhongtech.com") {
		return "kuaishou"
	}
	if host == "douyin.com" || strings.HasSuffix(host, ".douyin.com") ||
		host == "iesdouyin.com" || strings.HasSuffix(host, ".iesdouyin.com") {
		return "douyin"
	}
	if host == "xiaohongshu.com" || strings.HasSuffix(host, ".xiaohongshu.com") ||
		host == "xhslink.com" || strings.HasSuffix(host, ".xhslink.com") {
		return "xhs"
	}
	if host == "miyoushe.com" || strings.HasSuffix(host, ".miyoushe.com") {
		return "miyoushe"
	}
	return ""
}
