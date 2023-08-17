package hyliocache

import (
	"strconv"
	"strings"
)

// CheckAddr 检查addr是否为x.x.x.x:port的形式
func CheckAddr(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return false
	}
	ip, port := parts[0], parts[1]
	if _, err := strconv.Atoi(port); err != nil {
		return false
	}
	ipToken := strings.Split(ip, ".")
	if ip == "localhost" {
		return true
	}
	if len(ipToken) != 4 {
		return false
	}
	for _, x := range ipToken {
		if t, err := strconv.Atoi(x); t >= 256 || t < 0 || err != nil {
			return false
		}
	}
	return true
}
