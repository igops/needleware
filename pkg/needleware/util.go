package needleware

import (
	"net"
	"strconv"
)

func parseHostPort(addr string) (host string, port int32, err error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", -1, err
	}
	portInt, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return "", -1, err
	}
	return host, int32(portInt), nil
}
