package fortimgr

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// toString converts interface{} to string.
func toString(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case []any:
		if len(val) > 0 {
			return toString(val[0])
		}
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toStringSlice converts interface{} to []string.
func toStringSlice(v any) []string {
	if v == nil {
		return []string{}
	}
	switch val := v.(type) {
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = toString(item)
		}
		return result
	case []string:
		return val
	case string:
		if val == "" {
			return []string{}
		}
		return []string{val}
	default:
		return []string{fmt.Sprintf("%v", val)}
	}
}

// formatSubnet formats subnet from FortiManager.
// Input: string "ip/mask" or []any{"ip", "mask"}.
// Output: CIDR notation (e.g. "10.0.0.0/24"), host IP without /32.
func formatSubnet(v any) string {
	if v == nil {
		return ""
	}
	var rawSubnet string
	switch val := v.(type) {
	case string:
		rawSubnet = val
	case []any:
		if len(val) == 2 {
			rawSubnet = fmt.Sprintf("%s/%s", toString(val[0]), toString(val[1]))
		} else {
			return toString(val)
		}
	default:
		return fmt.Sprintf("%v", val)
	}
	return convertToCIDR(rawSubnet)
}

// convertToCIDR converts dotted mask to CIDR prefix.
// "192.168.1.0/255.255.255.0" → "192.168.1.0/24"
// "10.0.0.1/255.255.255.255" → "10.0.0.1" (strip /32)
// Already CIDR → pass through.
func convertToCIDR(subnet string) string {
	if subnet == "" {
		return ""
	}

	if !strings.Contains(subnet, "/") {
		return subnet
	}

	parts := strings.SplitN(subnet, "/", 2)
	if len(parts) != 2 {
		return subnet
	}

	ip := parts[0]
	mask := parts[1]

	// Already CIDR notation (numeric prefix).
	if prefix, err := strconv.Atoi(mask); err == nil {
		if prefix == 32 {
			return ip
		}
		return subnet
	}

	// Dotted mask notation — convert.
	prefix := maskToCIDRPrefix(mask)
	if prefix < 0 {
		return subnet
	}
	if prefix == 32 {
		return ip
	}
	return fmt.Sprintf("%s/%d", ip, prefix)
}

// maskToCIDRPrefix converts dotted mask string to prefix length.
// "255.255.255.0" → 24, "255.255.255.255" → 32.
// Returns -1 for invalid masks.
func maskToCIDRPrefix(mask string) int {
	ip := net.ParseIP(mask)
	if ip == nil {
		return -1
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return -1
	}
	ones, _ := net.IPv4Mask(ip4[0], ip4[1], ip4[2], ip4[3]).Size()
	return ones
}

// Enum maps — numeric strings from FlatUI API → official FortiOS names.
// If the value is already a named string, mapEnum passes it through unchanged.

var addressTypes = map[string]string{
	"0": "ipmask", "1": "iprange", "2": "fqdn", "3": "wildcard",
	"4": "geography", "5": "wildcard-fqdn", "6": "dynamic",
}

var policyActions = map[string]string{
	"0": "deny", "1": "accept", "2": "ipsec",
}

var enableDisable = map[string]string{
	"0": "disable", "1": "enable",
}

var logTrafficModes = map[string]string{
	"0": "disable", "1": "utm", "2": "all",
}

var serviceProtocols = map[string]string{
	"1": "ICMP", "2": "IP", "5": "TCP/UDP/SCTP", "6": "ICMP6",
}

var vipProtocols = map[string]string{
	"1": "tcp", "2": "udp", "3": "sctp", "4": "icmp",
}

var ippoolTypes = map[string]string{
	"0": "overload", "1": "one-to-one", "2": "fixed-port-range", "3": "port-block-allocation",
}

var deviceHAModes = map[string]string{
	"0": "standalone", "1": "master", "2": "slave",
}

var deviceConnStates = map[string]string{
	"0": "offline", "1": "online",
}

var adomStates = map[string]string{
	"0": "disabled", "1": "enabled",
}

var adomModes = map[string]string{
	"1": "normal", "2": "backup",
}

// mapEnum maps a numeric string to a named value.
// Named strings (non-numeric) pass through unchanged.
func mapEnum(v string, m map[string]string) string {
	if mapped, ok := m[v]; ok {
		return mapped
	}
	return v
}

// dayNames maps bit positions to day names (bit 0 = sunday).
var dayNames = [7]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

// mapScheduleDay converts a bitmask integer to space-separated day names.
// "127" → "sunday monday tuesday wednesday thursday friday saturday"
// Non-numeric strings pass through unchanged.
func mapScheduleDay(v string) string {
	n, err := strconv.Atoi(v)
	if err != nil {
		return v
	}
	var days []string
	for i, name := range dayNames {
		if n&(1<<uint(i)) != 0 {
			days = append(days, name)
		}
	}
	if len(days) == 0 {
		return "none"
	}
	return strings.Join(days, " ")
}

// formatMappedIP formats Virtual IP mappedip field.
// Can be string or []any of ranges → comma-joined.
func formatMappedIP(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = toString(item)
		}
		return strings.Join(parts, ",")
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatScheduleTime formats schedule time field.
// Can be string or []any like ["15:15", "2023/04/05"] → space-joined.
func formatScheduleTime(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []any:
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = toString(item)
		}
		return strings.Join(parts, " ")
	default:
		return fmt.Sprintf("%v", val)
	}
}
