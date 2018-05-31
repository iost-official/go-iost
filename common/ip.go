package common

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

var PublicIPDetechUrls = []string{
	"http://ipecho.net/plain",
	"http://myexternalip.com/raw",
}

func GetPulicIP() string {
	for _, detectUrl := range PublicIPDetechUrls {
		resp, err := http.Get(detectUrl)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		return strings.Trim(string(body), "\n")
	}
	return ""
}

func IsPublicIP(IP net.IP) bool {
	if IP.IsLoopback() || IP.IsLinkLocalMulticast() || IP.IsLinkLocalUnicast() {
		return false
	}
	if ip4 := IP.To4(); ip4 != nil {
		switch true {
		case ip4[0] == 10:
			return false
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return false
		case ip4[0] == 192 && ip4[1] == 168:
			return false
		default:
			return true
		}
	}
	return false
}
