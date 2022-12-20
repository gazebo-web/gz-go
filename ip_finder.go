package gz

// Code originally taken from
// https://husobee.github.io/golang/ip-address/2015/12/17/remote-ip-go.html

import (
	"bytes"
	"net"
	"net/http"
	"strings"
)

// ipRange is a structure that holds the start and end of a range of
// ip addresses
type ipRange struct {
	start net.IP
	end   net.IP
}

// private ranges lists IP ranges that are considered private or local
// addresses
var privateRanges = []ipRange{
	{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}

// inRange checks to see if a given ip address is within a range given
func inRange(r ipRange, ipAddress net.IP) bool {
	// strcmp type byte comparison
	if bytes.Compare(ipAddress, r.start) >= 0 &&
		bytes.Compare(ipAddress, r.end) < 0 {
		return true
	}
	return false
}

// isPrivateSubnet checks to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
	// my use case is only concerned with ipv4 atm
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		// iterate over all our ranges
		for _, r := range privateRanges {
			// check if this ip is in a private range
			if inRange(r, ipAddress) {
				return true
			}
		}
	}
	return false
}

// ///////////////////////////////////////////////
// / getIPAddress searches, from right to left, for a valid IP address in a
// / request.
func getIPAddress(r *http.Request) string {

	// Search over possible headers.
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {

		// Number of addresses in the header
		addresses := strings.Split(r.Header.Get(h), ",")

		// March from right to left until we get a public address
		// that will be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])

			// header can contain spaces too, strip those out.
			realIP := net.ParseIP(ip)

			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				// bad address, go to next
				continue
			}
			return ip
		}
	}
	return ""
}
