package fortimgr

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
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

// confStatuses maps FortiManager's conf_status int to a named value.
// Tracks whether the device's running config matches what FMG has on file.
var confStatuses = map[string]string{
	"0": "unknown",
	"1": "insync",
	"2": "modified",
}

// devStatuses maps FortiManager's dev_status int to a named value.
// Tracks the operational state of the device from FMG's perspective.
// Values are defined in the FortiManager DVMDB schema.
var devStatuses = map[string]string{
	"0":  "none",
	"1":  "unknown",
	"2":  "checked_in",
	"3":  "in_progress",
	"4":  "installed",
	"5":  "aborted",
	"6":  "sched",
	"7":  "retry",
	"8":  "canceled",
	"9":  "pending",
	"10": "retrieved",
	"11": "changed_conf",
	"12": "sync_fail",
	"13": "timeout",
	"14": "rev_revert",
	"15": "auto_updated",
}

// haRoles maps ha_slave[*].role (0/1) to a named role. Unknown ints fall
// back to the raw string via mapEnum's passthrough behavior.
var haRoles = map[string]string{
	"0": "slave",
	"1": "master",
}

// workflowStates maps FortiManager workflow session state ints to named
// values. The complete enum is not documented by FortiManager; only the
// "approved" value (3) has been observed empirically where sessions had
// create/submit/audit timestamps all populated. Unmapped ints pass
// through unchanged via mapEnum.
var objrevActions = map[string]string{
	"1": "add",
	"3": "modify",
}

var workflowStates = map[string]string{
	"3": "approved",
}

var adomStates = map[string]string{
	"0": "disabled", "1": "enabled",
}

var adomModes = map[string]string{
	"1": "normal", "2": "backup",
}

var vdomOpModes = map[string]string{
	"0": "nat", "1": "transparent",
}

var interfaceTypes = map[string]string{
	"0": "physical", "1": "vlan", "2": "aggregate", "3": "redundant",
	"4": "tunnel", "5": "wireless", "6": "vdom-link", "7": "loopback",
}

var interfaceStatuses = map[string]string{
	"0": "down", "1": "up",
}

var userTypes = map[string]string{
	"1": "local", "2": "radius", "3": "tacacs+", "4": "ldap",
}

var userGroupTypes = map[string]string{
	"0": "firewall", "1": "fsso-service", "2": "rsso", "3": "guest",
}

var ldapTypes = map[string]string{
	"0": "simple", "1": "anonymous", "2": "regular",
}

var ldapSecure = map[string]string{
	"0": "disable", "1": "starttls", "2": "ldaps",
}

var radiusAuthTypes = map[string]string{
	"0": "auto", "1": "ms_chap_v2", "2": "ms_chap", "3": "chap", "4": "pap",
}

var ipsecModes = map[string]string{
	"0": "main", "1": "aggressive",
}

var ipsecTypes = map[string]string{
	"0": "static", "1": "dynamic", "2": "ddns",
}

var scanModes = map[string]string{
	"0": "default", "1": "legacy", "2": "full",
}

var featureSets = map[string]string{
	"0": "flow", "1": "proxy",
}

var botnetConnections = map[string]string{
	"0": "disable", "1": "block", "2": "monitor",
}

var unknownAppActions = map[string]string{
	"0": "pass", "1": "block",
}

var inspectionModes = map[string]string{
	"0": "proxy", "1": "flow-based", "2": "dns",
}

var serverCertModes = map[string]string{
	"0": "re-sign", "1": "replace",
}

var supportedALPN = map[string]string{
	"0": "none", "1": "http1-1", "2": "http2", "3": "all",
}

var intrazoneTraffic = map[string]string{
	"0": "allow", "1": "deny",
}

var interfaceRoles = map[string]string{
	"0": "lan", "1": "wan", "2": "dmz", "3": "undefined",
}

var interfaceModes = map[string]string{
	"0": "static", "1": "dhcp", "2": "pppoe",
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

// unixToTime converts a FortiManager Unix timestamp to time.Time.
// Returns the zero time.Time (IsZero() == true) when the value is nil,
// zero, or otherwise invalid — FortiManager uses 0 to mean "never".
func unixToTime(v any) time.Time {
	if v == nil {
		return time.Time{}
	}
	var sec int64
	switch val := v.(type) {
	case int:
		sec = int64(val)
	case int64:
		sec = val
	case float64:
		sec = int64(val)
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return time.Time{}
		}
		sec = n
	default:
		return time.Time{}
	}
	if sec <= 0 {
		return time.Time{}
	}
	return time.Unix(sec, 0).UTC()
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
