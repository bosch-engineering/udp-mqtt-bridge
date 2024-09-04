package utils

import (
	"errors"
	"net"

	"github.com/gookit/slog"
)

func GetBroadcastAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		slog.Errorf("Error getting network interfaces: %v", err)
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagBroadcast != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				slog.Errorf("Error getting addresses for interface %s: %v", iface.Name, err)
				return "", err
			}

			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					return ipNet.IP.Mask(ipNet.IP.DefaultMask()).String(), nil
				}
			}
		}
	}

	return "", errors.New("no broadcast address found")
}
