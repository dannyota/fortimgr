package fortimgr

import (
	"reflect"
	"testing"
)

func TestDecodeAllowAccess(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want []string
	}{
		// Numeric bitmasks (the default FortiManager API form) — verified observed values.
		{"none", float64(0), []string{}},
		{"https only", float64(2), []string{"https"}},
		{"classic mgmt set", float64(15), []string{"ping", "https", "ssh", "snmp"}},
		{"https + fabric probe", float64(1026), []string{"https", "probe-response"}},
		{"fmg-managed iface", float64(1175), []string{"ping", "https", "ssh", "http", "fgfm", "probe-response"}},
		// Reference verbose dump 50879 must round-trip exactly, in ascending bit order.
		{"reference 50879", float64(50879), []string{
			"ping", "https", "ssh", "snmp", "http", "telnet", "fgfm",
			"radius-acct", "probe-response", "fabric", "speed-test",
		}},
		// Other accepted shapes.
		{"numeric string", "15", []string{"ping", "https", "ssh", "snmp"}},
		{"symbolic string", "ping https ssh", []string{"ping", "https", "ssh"}},
		{"verbose array", []any{"ping", "https"}, []string{"ping", "https"}},
		{"nil", nil, []string{}},
		{"empty string", "", []string{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := decodeAllowAccess(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("decodeAllowAccess(%v) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestDecodeAllowAccessBitsNoLeftover(t *testing.T) {
	// Every observed mask must be fully accounted for (sum of decoded bits == mask).
	for _, mask := range []int64{0, 2, 15, 1026, 1175, 50879} {
		var sum int64
		for _, b := range allowAccessBits {
			if mask&b.bit != 0 {
				sum |= b.bit
			}
		}
		if sum != mask {
			t.Errorf("mask %d has unaccounted bits: decoded %d", mask, sum)
		}
	}
}
