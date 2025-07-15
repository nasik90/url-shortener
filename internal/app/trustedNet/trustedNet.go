package trustednet

import (
	"errors"
	"net"
	"strings"

	"github.com/nasik90/url-shortener/cmd/shortener/settings"
)

// CheckForTrustedNet - проверяет IP адрес на то, что IP входит в доверенную сеть.
// Проверка происходит не для всех методов.
func CheckForTrustedNet(trustedSubnet, ipStr, method string) error {
	if !neededToCheckMethod(method) {
		return nil
	}
	// Провкра trustedSubnet из настроек сервиса
	if trustedSubnet == "" {
		return errors.New("trusted subnet is empty, access forbidden")
	}
	_, trustedNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return err
	}

	// Проверка ip
	if ipStr == "" {
		return errors.New("forbidden - missing X-Real-IP header")
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errors.New("forbidden - invalid IP")
	}

	if !trustedNet.Contains(ip) {
		return errors.New("forbidden - IP not in trusted subnet")
	}
	return nil
}

func neededToCheckMethod(method string) bool {

	needToCheck := false
	for _, methodToCheck := range settings.MethodsToCheckTrustedNet {
		if strings.Contains(method, methodToCheck.GRPSMethod) || strings.Contains(method, methodToCheck.APIMethod) {
			needToCheck = true
		}
	}

	return needToCheck

}
