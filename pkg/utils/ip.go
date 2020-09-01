package utils

import (
	"errors"
	"net"
)

func GetIPAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip.String() == "127.0.0.1" {
				continue
			}

			if checkIPv4(ip.String()) {
				return ip.String(), nil
			}
		}
	}
	return "", errors.New("did not find ip")
}

func checkIPv4(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	}
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			return true
		}
	}
	return false
}
