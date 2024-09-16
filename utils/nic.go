package utils

import (
	"errors"
	"fmt"
	"net"
)

var (
	errPrivateIPv4FailedToDerive = errors.New("failed to derive private ipv4")
)

func PrivateIPv4() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("%w: %w",
			errPrivateIPv4FailedToDerive, err,
		)
	}

	for _, ifs := range interfaces {
		addrs, err := ifs.Addrs()
		if err != nil {
			return nil, fmt.Errorf("%w: %w",
				errPrivateIPv4FailedToDerive, err,
			)
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ipv4 := ipNet.IP.To4()
			if ipv4 == nil {
				continue
			}

			if !ipv4.IsPrivate() {
				continue
			}

			return ipv4, nil
		}
	}

	return nil, fmt.Errorf("%w: found no ipv4 interfaces",
		errPrivateIPv4FailedToDerive,
	)
}
