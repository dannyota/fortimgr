package fortimgr

import (
	"strconv"
	"strings"
)

// FortiManager's JSON API returns the interface `allowaccess` administrative-access option
// as an INTEGER BITMASK (a FortiOS bitfield), not as protocol names — unlike a direct
// FortiGate, which returns the symbolic list. decodeAllowAccess turns that bitmask into the
// protocol names the SDK's Interface.AllowAccess promises, while still accepting an already-
// symbolic value (a string list, or a "verbose":1 array) unchanged.
//
// Bit values are the modern FortiOS ordering (ping = bit 0). Verified against the FortiManager
// API verbose dump 50879 = ping https ssh snmp http telnet fgfm radius-acct probe-response
// fabric speed-test, and decodes real observed masks (0, 2, 15, 1026, 1175) with no leftover
// bits. Note: bit 6 (value 64) is reserved/unassigned — fgfm is bit 7 (128).
var allowAccessBits = []struct {
	bit  int64
	name string
}{
	{1, "ping"},
	{2, "https"},
	{4, "ssh"},
	{8, "snmp"},
	{16, "http"},
	{32, "telnet"},
	{128, "fgfm"},
	{256, "auto-ipsec"},
	{512, "radius-acct"},
	{1024, "probe-response"},
	{2048, "capwap"},
	{4096, "dnp"},
	{8192, "ftm"},
	{16384, "fabric"},
	{32768, "speed-test"},
	{65536, "icond"},
	{131072, "scim"},
}

// decodeAllowAccessBits returns the protocol names enabled in a FortiOS allowaccess bitmask,
// in ascending bit order.
func decodeAllowAccessBits(mask int64) []string {
	out := make([]string, 0, len(allowAccessBits))
	for _, b := range allowAccessBits {
		if mask&b.bit != 0 {
			out = append(out, b.name)
		}
	}
	return out
}

// decodeAllowAccess normalizes the raw `allowaccess` value (numeric bitmask, numeric string,
// already-symbolic string, or string array) into protocol names.
func decodeAllowAccess(v any) []string {
	switch val := v.(type) {
	case nil:
		return []string{}
	case float64:
		return decodeAllowAccessBits(int64(val))
	case int:
		return decodeAllowAccessBits(int64(val))
	case int64:
		return decodeAllowAccessBits(val)
	case string:
		s := strings.TrimSpace(val)
		if s == "" {
			return []string{}
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return decodeAllowAccessBits(n)
		}
		return strings.Fields(s) // already symbolic, space-separated
	case []any:
		// "verbose":1 form — a list of symbolic names (or, defensively, numbers).
		out := make([]string, 0, len(val))
		for _, item := range val {
			out = append(out, decodeAllowAccess(item)...)
		}
		return out
	case []string:
		return val
	default:
		return []string{}
	}
}
