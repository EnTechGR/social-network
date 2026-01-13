// Add this to API/utils/validation.go or create API/utils/user_agent.go

package utils

import (
	"regexp"
	"strings"
)

// UserAgentInfo represents parsed User-Agent information
type UserAgentInfo struct {
	Browser        string // Chrome, Firefox, Safari, Edge, etc.
	BrowserVersion string // Major.Minor version
	OS             string // Windows, macOS, Linux, iOS, Android
	OSVersion      string // Version string
	DeviceType     string // Desktop, Mobile, Tablet
	IsBot          bool   // Whether this appears to be a bot
	Raw            string // Original User-Agent string
}

// ParseUserAgent extracts key information from User-Agent string
// This is intentionally simple - we don't need perfect parsing, just enough to detect major changes
func ParseUserAgent(userAgent string) UserAgentInfo {
	ua := UserAgentInfo{
		Raw: userAgent,
	}
	
	if userAgent == "" {
		return ua
	}
	
	lower := strings.ToLower(userAgent)
	
	// Detect bots (common patterns)
	botPatterns := []string{"bot", "crawl", "spider", "scraper", "curl", "wget"}
	for _, pattern := range botPatterns {
		if strings.Contains(lower, pattern) {
			ua.IsBot = true
			return ua
		}
	}
	
	// Browser detection (order matters - check most specific first)
	if strings.Contains(lower, "edg/") || strings.Contains(lower, "edge/") {
		ua.Browser = "Edge"
		ua.BrowserVersion = extractVersion(userAgent, `Edge?/(\d+\.\d+)`)
	} else if strings.Contains(lower, "chrome/") {
		ua.Browser = "Chrome"
		ua.BrowserVersion = extractVersion(userAgent, `Chrome/(\d+\.\d+)`)
	} else if strings.Contains(lower, "firefox/") {
		ua.Browser = "Firefox"
		ua.BrowserVersion = extractVersion(userAgent, `Firefox/(\d+\.\d+)`)
	} else if strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome") {
		ua.Browser = "Safari"
		ua.BrowserVersion = extractVersion(userAgent, `Version/(\d+\.\d+)`)
	} else if strings.Contains(lower, "opera") || strings.Contains(lower, "opr/") {
		ua.Browser = "Opera"
		ua.BrowserVersion = extractVersion(userAgent, `(?:Opera|OPR)/(\d+\.\d+)`)
	} else {
		ua.Browser = "Unknown"
	}
	
	// OS detection
	if strings.Contains(lower, "windows") {
		ua.OS = "Windows"
		if strings.Contains(lower, "windows nt 10") {
			ua.OSVersion = "10+"
		} else if strings.Contains(lower, "windows nt 6") {
			ua.OSVersion = "7/8"
		}
	} else if strings.Contains(lower, "mac os x") || strings.Contains(lower, "macos") {
		ua.OS = "macOS"
		ua.OSVersion = extractVersion(userAgent, `Mac OS X (\d+[._]\d+)`)
	} else if strings.Contains(lower, "linux") {
		ua.OS = "Linux"
	} else if strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") {
		ua.OS = "iOS"
		ua.OSVersion = extractVersion(userAgent, `OS (\d+[._]\d+)`)
	} else if strings.Contains(lower, "android") {
		ua.OS = "Android"
		ua.OSVersion = extractVersion(userAgent, `Android (\d+\.\d+)`)
	} else {
		ua.OS = "Unknown"
	}
	
	// Device type detection
	if strings.Contains(lower, "mobile") || strings.Contains(lower, "iphone") || strings.Contains(lower, "android") {
		if strings.Contains(lower, "tablet") || strings.Contains(lower, "ipad") {
			ua.DeviceType = "Tablet"
		} else {
			ua.DeviceType = "Mobile"
		}
	} else {
		ua.DeviceType = "Desktop"
	}
	
	return ua
}

// extractVersion extracts version number using regex pattern
func extractVersion(userAgent, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(userAgent)
	if len(matches) > 1 {
		// Return major.minor only (ignore patch versions)
		version := matches[1]
		parts := strings.SplitN(version, ".", 3)
		if len(parts) >= 2 {
			return parts[0] + "." + parts[1]
		}
		return parts[0]
	}
	return ""
}

// ValidateUserAgent checks if the current User-Agent is compatible with the stored one
// Returns: (isValid bool, suspicionLevel string, reason string)
// Suspicion levels: "none", "low", "medium", "high"
func ValidateUserAgent(stored, current string) (bool, string, string) {
	if stored == "" {
		// No stored User-Agent - this is the first request, consider valid
		return true, "none", ""
	}
	
	if current == "" {
		// Request has no User-Agent - suspicious but not necessarily malicious
		return true, "low", "Missing User-Agent in request"
	}
	
	// Exact match - completely normal
	if stored == current {
		return true, "none", ""
	}
	
	// Parse both User-Agents
	storedUA := ParseUserAgent(stored)
	currentUA := ParseUserAgent(current)
	
	// Bot detection
	if storedUA.IsBot || currentUA.IsBot {
		return true, "low", "Bot user agent detected"
	}
	
	// Critical mismatches - REJECT
	if storedUA.Browser != currentUA.Browser {
		// Different browser entirely - strong indicator of session hijacking
		return false, "high", "Browser changed from " + storedUA.Browser + " to " + currentUA.Browser
	}
	
	if storedUA.OS != currentUA.OS {
		// Different OS - strong indicator of session hijacking
		return false, "high", "Operating system changed from " + storedUA.OS + " to " + currentUA.OS
	}
	
	if storedUA.DeviceType != currentUA.DeviceType {
		// Desktop to Mobile or vice versa - suspicious
		return false, "high", "Device type changed from " + storedUA.DeviceType + " to " + currentUA.DeviceType
	}
	
	// Minor version changes - ALLOW (auto-updates are common)
	if storedUA.Browser == currentUA.Browser && storedUA.OS == currentUA.OS {
		// Same browser and OS, just version differences
		if storedUA.BrowserVersion != "" && currentUA.BrowserVersion != "" {
			if storedUA.BrowserVersion != currentUA.BrowserVersion {
				// Browser version changed (auto-update) - acceptable
				return true, "low", "Browser version updated from " + storedUA.BrowserVersion + " to " + currentUA.BrowserVersion
			}
		}
		
		if storedUA.OSVersion != "" && currentUA.OSVersion != "" {
			if storedUA.OSVersion != currentUA.OSVersion {
				// OS version changed (system update) - acceptable
				return true, "low", "OS version updated from " + storedUA.OSVersion + " to " + currentUA.OSVersion
			}
		}
		
		// Minor string differences (locale, build number, etc.) - acceptable
		return true, "low", "Minor User-Agent variation detected"
	}
	
	// Unknown mismatch
	return false, "medium", "User-Agent string changed unexpectedly"
}

// GetUserAgent safely extracts User-Agent from HTTP request
func GetUserAgent(headers map[string][]string) string {
	if ua, ok := headers["User-Agent"]; ok && len(ua) > 0 {
		return ua[0]
	}
	return ""
}