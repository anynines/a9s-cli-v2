package klutch

import (
	"errors"
	"net"
)

var (
	errNoSuitableIp = errors.New("no suitable local IP address found")
)

// determineHostLocalIP retrieves a local IP address of the host to be used for the Control Plane Cluster.
// It only considers ipv4 IPs from private ranges.
// TODO: is a ipv4 address always guaranteed in the local network?
// TODO: inform users that a network change will cause their demo to break
// TODO: listening on this IP implies that the kind cluster and ingress will be reachable by anyone in the user's local network.
// Find/review alternative solutions to avoid this security risk
func determineHostLocalIP() (string, error) {
	return "host.docker.internal", nil

	// ifAddrs, err := net.InterfaceAddrs()
	// if err != nil {
	// 	return "", err
	// }

	// return getFirstPrivateIPv4(ifAddrs)
}

func getFirstPrivateIPv4(ifAddrs []net.Addr) (string, error) {
	for _, addr := range ifAddrs {
		ipnet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		if ipnet.IP.IsLoopback() {
			continue
		}

		if !ipnet.IP.IsPrivate() {
			continue
		}

		if ipnet.IP.To4() != nil {
			return ipnet.IP.String(), nil
		}
	}

	return "", errNoSuitableIp
}
