package klutch

import (
	"errors"
	"net"
	"testing"
)

var mockInterfaceAddrs = []net.Addr{
	// IPv4 loopback
	&net.IPNet{
		IP:   net.IPv4(127, 0, 0, 1),
		Mask: net.CIDRMask(8, 32),
	},
	&net.IPNet{
		IP:   net.IPv4(192, 168, 1, 100),
		Mask: net.CIDRMask(24, 32),
	},
	&net.IPNet{
		IP:   net.IPv4(8, 8, 8, 8),
		Mask: net.CIDRMask(24, 32),
	},
	// IPv6 loopback
	&net.IPNet{
		IP:   net.ParseIP("::1"),
		Mask: net.CIDRMask(128, 128),
	},
	&net.IPNet{
		IP:   net.ParseIP("2001:db8::1"),
		Mask: net.CIDRMask(64, 128),
	},
	&net.IPNet{
		IP:   net.ParseIP("fe80::1"),
		Mask: net.CIDRMask(64, 128),
	},
}

var mockInterfaceAddrsMultipleSuitable = []net.Addr{
	// IPv4 loopback
	&net.IPNet{
		IP:   net.IPv4(127, 0, 0, 1),
		Mask: net.CIDRMask(8, 32),
	},
	&net.IPNet{
		IP:   net.IPv4(172, 16, 0, 1),
		Mask: net.CIDRMask(24, 32),
	},
	&net.IPNet{
		IP:   net.IPv4(192, 168, 1, 100),
		Mask: net.CIDRMask(24, 32),
	},
	&net.IPNet{
		IP:   net.IPv4(8, 8, 8, 8),
		Mask: net.CIDRMask(24, 32),
	},
	// IPv6 loopback
	&net.IPNet{
		IP:   net.ParseIP("::1"),
		Mask: net.CIDRMask(128, 128),
	},
	&net.IPNet{
		IP:   net.ParseIP("2001:db8::1"),
		Mask: net.CIDRMask(64, 128),
	},
	&net.IPNet{
		IP:   net.ParseIP("fe80::1"),
		Mask: net.CIDRMask(64, 128),
	},
}

var mockInterfaceAddrsNoSuitable = []net.Addr{
	&net.IPNet{
		IP:   net.IPv4(127, 0, 0, 1),
		Mask: net.CIDRMask(8, 32),
	},
	&net.IPNet{
		IP:   net.IPv4(8, 8, 8, 8),
		Mask: net.CIDRMask(24, 32),
	},
	&net.IPNet{
		IP:   net.ParseIP("::1"),
		Mask: net.CIDRMask(128, 128),
	},
	&net.IPNet{
		IP:   net.ParseIP("2001:db8::1"),
		Mask: net.CIDRMask(64, 128),
	},
	&net.IPNet{
		IP:   net.ParseIP("fe80::1"),
		Mask: net.CIDRMask(64, 128),
	},
}

var mockInterfaceAddrsEmpty = []net.Addr{}

func TestGetFirstPrivateIPv4(t *testing.T) {
	testCases := []struct {
		name          string
		ipAddrMock    []net.Addr
		expectedIp    string
		expectedError error
	}{
		{
			name:          "positive",
			ipAddrMock:    mockInterfaceAddrs,
			expectedIp:    "192.168.1.100",
			expectedError: nil,
		},
		{
			name:          "positive multiple suitable",
			ipAddrMock:    mockInterfaceAddrsMultipleSuitable,
			expectedIp:    "172.16.0.1",
			expectedError: nil,
		},
		{
			name:          "error no suitable ip",
			ipAddrMock:    mockInterfaceAddrsNoSuitable,
			expectedIp:    "",
			expectedError: errNoSuitableIp,
		},
		{
			name:          "error no suitable ip empty",
			ipAddrMock:    mockInterfaceAddrsEmpty,
			expectedIp:    "",
			expectedError: errNoSuitableIp,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ip, err := getFirstPrivateIPv4(tc.ipAddrMock)
			if tc.expectedError != nil && !errors.Is(err, tc.expectedError) {
				t.Fatalf("expected error %v, but got %v", tc.expectedError, err)
			}

			if ip != tc.expectedIp {
				t.Fatalf("expected ip %s, but got %s", tc.expectedIp, ip)
			}
		})
	}
}
