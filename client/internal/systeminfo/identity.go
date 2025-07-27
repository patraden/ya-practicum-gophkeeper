package clientinfo

import (
	"fmt"
	"net"
	"os"
	"runtime"
)

// GenerateClientInfo returns a stable, human-readable ID like:
// "mac=00:11:22:33:44:55;host=macbook-pro;ip=192.168.1.5;os=darwin".
func GenerateClientInfo() string {
	mac := getMACAddress()
	hostname := getHostname()
	ip := getLocalIP()
	osName := runtime.GOOS

	return fmt.Sprintf("mac=%s;host=%s;ip=%s;os=%s", mac, hostname, ip, osName)
}

func getMACAddress() string {
	ifs, _ := net.Interfaces()
	for _, iface := range ifs {
		if len(iface.HardwareAddr) == 0 {
			continue
		}

		return iface.HardwareAddr.String()
	}

	return "unknown-mac"
}

func getHostname() string {
	name, err := os.Hostname()
	if err != nil || name == "" {
		return "unknown-host"
	}

	return name
}

//nolint:varnamelen // reason : ip is not confusing at all.
func getLocalIP() string {
	conns, _ := net.Interfaces()
	for _, iface := range conns {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip != nil && ip.IsPrivate() && ip.To4() != nil {
				return ip.String()
			}
		}
	}

	return "unknown-ip"
}
