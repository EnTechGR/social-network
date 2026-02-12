// Add this to API/utils/validation.go or create API/utils/ip_validation.go

package utils

import (
	"net"
	"strings"
)

// IPValidationMode defines how strictly to validate IP address changes
type IPValidationMode int

const (
	// IPValidationStrict requires exact IP match (not recommended for production)
	IPValidationStrict IPValidationMode = iota
	
	// IPValidationSubnet allows IPs within same /24 subnet (recommended for most apps)
	// Good balance: detects cross-network attacks while allowing DHCP/mobile changes
	IPValidationSubnet
	
	// IPValidationLenient logs changes but doesn't block (monitoring mode)
	IPValidationLenient
)

// Current validation mode - configure based on your security requirements
// For social networks: Subnet matching is recommended
var CurrentIPValidationMode = IPValidationLenient

// ValidateIPAddress checks if the current IP address is acceptable given the stored IP
// Returns: (isValid bool, suspicionLevel string, reason string)
// Suspicion levels: "none", "low", "medium", "high"
func ValidateIPAddress(storedIP, currentIP string) (bool, string, string) {
	// Handle empty cases
	if storedIP == "" {
		// No stored IP - first request, consider valid
		return true, "none", ""
	}
	
	if currentIP == "" {
		// No current IP - suspicious but might be proxy/load balancer issue
		return true, "medium", "Missing IP address in request"
	}
	
	// Clean IPs (remove port if present)
	storedIP = extractIP(storedIP)
	currentIP = extractIP(currentIP)
	
	// Exact match - completely normal
	if storedIP == currentIP {
		return true, "none", ""
	}
	
	// Different IP - level of concern depends on validation mode
	switch CurrentIPValidationMode {
	case IPValidationStrict:
		// Strict mode: any IP change is rejected
		return false, "high", "IP address changed (strict mode)"
		
	case IPValidationSubnet:
		// Subnet mode: check if IPs are in same /24 subnet
		if inSameSubnet(storedIP, currentIP) {
			return true, "low", "IP changed within same subnet (DHCP/mobile network)"
		}
		// Different subnet - suspicious
		return false, "high", "IP changed to different subnet/network"
		
	case IPValidationLenient:
		// Lenient mode: log but allow all changes
		return true, "medium", "IP address changed (lenient mode)"
		
	default:
		// Default to subnet validation
		if inSameSubnet(storedIP, currentIP) {
			return true, "low", "IP changed within same subnet"
		}
		return false, "high", "IP changed to different subnet"
	}
}

// extractIP removes port from IP:port format
// Example: "192.168.1.50:12345" -> "192.168.1.50"
func extractIP(ipWithPort string) string {
	// Handle IPv6 addresses in brackets
	if strings.HasPrefix(ipWithPort, "[") {
		// [::1]:8080 -> ::1
		endBracket := strings.Index(ipWithPort, "]")
		if endBracket > 0 {
			return ipWithPort[1:endBracket]
		}
	}
	
	// Handle IPv4 with port
	if colonIndex := strings.LastIndex(ipWithPort, ":"); colonIndex > 0 {
		// Check if this is actually a port (not part of IPv6)
		potentialIP := ipWithPort[:colonIndex]
		if net.ParseIP(potentialIP) != nil {
			return potentialIP
		}
	}
	
	return ipWithPort
}

// inSameSubnet checks if two IPs are in the same /24 subnet (IPv4) or /64 subnet (IPv6)
func inSameSubnet(ip1, ip2 string) bool {
	parsedIP1 := net.ParseIP(ip1)
	parsedIP2 := net.ParseIP(ip2)
	
	if parsedIP1 == nil || parsedIP2 == nil {
		// Can't parse one or both IPs - consider different for safety
		return false
	}
	
	// Check if both are IPv4 or both are IPv6
	isIPv4_1 := parsedIP1.To4() != nil
	isIPv4_2 := parsedIP2.To4() != nil
	
	if isIPv4_1 != isIPv4_2 {
		// One is IPv4, other is IPv6 - definitely different
		return false
	}
	
	if isIPv4_1 {
		// Both IPv4 - check /24 subnet (Class C)
		// Example: 192.168.1.x are all in same subnet
		_, subnet1, _ := net.ParseCIDR(ip1 + "/24")
		_, subnet2, _ := net.ParseCIDR(ip2 + "/24")
		
		if subnet1 == nil || subnet2 == nil {
			return false
		}
		
		// Check if IP2 is in subnet1 and IP1 is in subnet2
		return subnet1.Contains(parsedIP2) && subnet2.Contains(parsedIP1)
	} else {
		// Both IPv6 - check /64 subnet
		// IPv6 /64 is standard for a single network segment
		_, subnet1, _ := net.ParseCIDR(ip1 + "/64")
		_, subnet2, _ := net.ParseCIDR(ip2 + "/64")
		
		if subnet1 == nil || subnet2 == nil {
			return false
		}
		
		return subnet1.Contains(parsedIP2) && subnet2.Contains(parsedIP1)
	}
}

// GetClientIP extracts the real client IP from the request, handling proxies and load balancers
// Checks X-Forwarded-For, X-Real-IP headers first, falls back to RemoteAddr
func GetClientIP(remoteAddr string, headers map[string][]string) string {
	// Priority order:
	// 1. X-Forwarded-For (most common proxy header)
	// 2. X-Real-IP (nginx proxy)
	// 3. RemoteAddr (direct connection)
	
	// Check X-Forwarded-For (can have multiple IPs: "client, proxy1, proxy2")
	if xff, ok := headers["X-Forwarded-For"]; ok && len(xff) > 0 && xff[0] != "" {
		// Take the first IP (original client)
		ips := strings.Split(xff[0], ",")
		if len(ips) > 0 {
			clientIP := strings.TrimSpace(ips[0])
			if clientIP != "" {
				return extractIP(clientIP)
			}
		}
	}
	
	// Check X-Real-IP
	if xri, ok := headers["X-Real-Ip"]; ok && len(xri) > 0 && xri[0] != "" {
		return extractIP(xri[0])
	}
	
	// Fall back to RemoteAddr
	return extractIP(remoteAddr)
}

// IsPrivateIP checks if an IP address is private/internal
// Useful for identifying requests from internal networks
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(extractIP(ipStr))
	if ip == nil {
		return false
	}
	
	// Check common private ranges
	privateRanges := []string{
		"10.0.0.0/8",        // Class A private
		"172.16.0.0/12",     // Class B private
		"192.168.0.0/16",    // Class C private
		"127.0.0.0/8",       // Loopback
		"169.254.0.0/16",    // Link-local
		"::1/128",           // IPv6 loopback
		"fe80::/10",         // IPv6 link-local
		"fc00::/7",          // IPv6 unique local
	}
	
	for _, cidr := range privateRanges {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return true
		}
	}
	
	return false
}

// IPInfo represents detailed IP address information for logging
type IPInfo struct {
	IP        string
	IsIPv4    bool
	IsIPv6    bool
	IsPrivate bool
	Subnet24  string // /24 for IPv4, /64 for IPv6
}

// AnalyzeIP provides detailed information about an IP address for security logging
func AnalyzeIP(ipStr string) IPInfo {
	ipStr = extractIP(ipStr)
	ip := net.ParseIP(ipStr)
	
	info := IPInfo{
		IP: ipStr,
	}
	
	if ip == nil {
		return info
	}
	
	info.IsIPv4 = ip.To4() != nil
	info.IsIPv6 = !info.IsIPv4
	info.IsPrivate = IsPrivateIP(ipStr)
	
	// Determine subnet
	if info.IsIPv4 {
		_, subnet, _ := net.ParseCIDR(ipStr + "/24")
		if subnet != nil {
			info.Subnet24 = subnet.String()
		}
	} else {
		_, subnet, _ := net.ParseCIDR(ipStr + "/64")
		if subnet != nil {
			info.Subnet24 = subnet.String()
		}
	}
	
	return info
}