package fortimgr

import "testing"

func TestToString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"float64", float64(42), "42"},
		{"float64 decimal", float64(3.14), "3"},
		{"int", 7, "7"},
		{"slice first", []any{"first", "second"}, "first"},
		{"slice empty", []any{}, ""},
		{"slice nested float", []any{float64(99)}, "99"},
		{"bool fallback", true, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.in)
			if got != tt.want {
				t.Errorf("toString(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want []string
	}{
		{"nil", nil, []string{}},
		{"[]any", []any{"a", "b"}, []string{"a", "b"}},
		{"[]any with float", []any{"x", float64(1)}, []string{"x", "1"}},
		{"[]string", []string{"p", "q"}, []string{"p", "q"}},
		{"string", "single", []string{"single"}},
		{"empty string", "", []string{}},
		{"fallback", 42, []string{"42"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSlice(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFormatSubnet(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"cidr string", "10.0.0.0/24", "10.0.0.0/24"},
		{"dotted mask string", "192.168.1.0/255.255.255.0", "192.168.1.0/24"},
		{"host /32 string", "10.0.0.1/255.255.255.255", "10.0.0.1"},
		{"host /32 numeric", "10.0.0.1/32", "10.0.0.1"},
		{"no slash", "10.0.0.1", "10.0.0.1"},
		{"array ip mask", []any{"172.16.0.0", "255.255.0.0"}, "172.16.0.0/16"},
		{"array host", []any{"1.2.3.4", "255.255.255.255"}, "1.2.3.4"},
		{"array single", []any{"hello"}, "hello"},
		{"fallback", 12345, "12345"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSubnet(tt.in)
			if got != tt.want {
				t.Errorf("formatSubnet(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestConvertToCIDR(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"10.0.0.1", "10.0.0.1"},
		{"10.0.0.0/24", "10.0.0.0/24"},
		{"10.0.0.0/8", "10.0.0.0/8"},
		{"10.0.0.1/32", "10.0.0.1"},
		{"192.168.1.0/255.255.255.0", "192.168.1.0/24"},
		{"10.0.0.0/255.0.0.0", "10.0.0.0/8"},
		{"10.0.0.1/255.255.255.255", "10.0.0.1"},
		{"bad/mask", "bad/mask"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := convertToCIDR(tt.in)
			if got != tt.want {
				t.Errorf("convertToCIDR(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaskToCIDRPrefix(t *testing.T) {
	tests := []struct {
		mask string
		want int
	}{
		{"255.255.255.255", 32},
		{"255.255.255.0", 24},
		{"255.255.0.0", 16},
		{"255.0.0.0", 8},
		{"0.0.0.0", 0},
		{"invalid", -1},
		{"", -1},
	}
	for _, tt := range tests {
		t.Run(tt.mask, func(t *testing.T) {
			got := maskToCIDRPrefix(tt.mask)
			if got != tt.want {
				t.Errorf("maskToCIDRPrefix(%q) = %d, want %d", tt.mask, got, tt.want)
			}
		})
	}
}

func TestFormatMappedIP(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"string", "10.0.0.1-10.0.0.5", "10.0.0.1-10.0.0.5"},
		{"array", []any{"10.0.0.1-10.0.0.5", "10.0.1.1-10.0.1.5"}, "10.0.0.1-10.0.0.5,10.0.1.1-10.0.1.5"},
		{"array single", []any{"10.0.0.1"}, "10.0.0.1"},
		{"fallback", 42, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMappedIP(tt.in)
			if got != tt.want {
				t.Errorf("formatMappedIP(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMapEnum(t *testing.T) {
	tests := []struct {
		name string
		v    string
		m    map[string]string
		want string
	}{
		{"mapped", "0", map[string]string{"0": "deny", "1": "accept"}, "deny"},
		{"mapped 1", "1", map[string]string{"0": "deny", "1": "accept"}, "accept"},
		{"passthrough string", "accept", map[string]string{"0": "deny", "1": "accept"}, "accept"},
		{"unknown key", "99", map[string]string{"0": "deny"}, "99"},
		{"empty string", "", map[string]string{"0": "deny"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapEnum(tt.v, tt.m)
			if got != tt.want {
				t.Errorf("mapEnum(%q) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestMapScheduleDay(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"all days", "127", "sunday monday tuesday wednesday thursday friday saturday"},
		{"weekdays", "62", "monday tuesday wednesday thursday friday"},
		{"weekend", "65", "sunday saturday"},
		{"sunday only", "1", "sunday"},
		{"monday only", "2", "monday"},
		{"saturday only", "64", "saturday"},
		{"zero", "0", "none"},
		{"string passthrough", "monday wednesday", "monday wednesday"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapScheduleDay(tt.in)
			if got != tt.want {
				t.Errorf("mapScheduleDay(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatScheduleTime(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"string", "08:00", "08:00"},
		{"array", []any{"15:15", "2023/04/05"}, "15:15 2023/04/05"},
		{"array single", []any{"12:00"}, "12:00"},
		{"fallback", 42, "42"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatScheduleTime(tt.in)
			if got != tt.want {
				t.Errorf("formatScheduleTime(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
