package fortimgr

import (
	"context"
	"fmt"
)

type apiNormalizedInterface struct {
	Name           string                  `json:"name"`
	OID            int                     `json:"oid"`
	SingleIntf     int                     `json:"single-intf"`
	ZoneOnly       int                     `json:"zone-only"`
	Wildcard       int                     `json:"wildcard"`
	DefaultMapping int                     `json:"default-mapping"`
	Color          int                     `json:"color"`
	DynamicMapping []apiDynamicIntfMapping `json:"dynamic_mapping"`
}

type apiDynamicIntfMapping struct {
	Scope         []apiDynamicIntfScope `json:"_scope"`
	LocalIntf     []string              `json:"local-intf"`
	IntrazoneDeny any                   `json:"intrazone-deny"`
}

type apiDynamicIntfScope struct {
	Name string `json:"name"`
	VDOM string `json:"vdom"`
}

// ListNormalizedInterfaces returns every normalized interface defined
// at the ADOM level. Normalized interfaces are FortiManager's per-ADOM
// interface abstraction: policies reference a normalized name like
// "wan1" or "internal", and FortiManager rewrites the policy per-device
// based on the Mappings defined on each normalized interface.
//
// Without this abstraction, a policy authored once would only work on
// devices whose physical interfaces happen to share the same name.
// With it, the same policy applies across a heterogeneous fleet.
//
// The SDK flattens the raw dynamic_mapping[] array: each _scope element
// becomes its own NormalizedInterfaceMapping entry. A normalized
// interface with a _scope of three {device, vdom} pairs therefore
// produces three Mappings entries — one per scope — making downstream
// iteration straightforward.
//
// Most normalized interfaces on a typical ADOM are unmapped
// declarations (no Mappings entries). Only the ones that policies
// actually reference per-device-per-vdom carry populated Mappings.
//
// Pagination is applied transparently (1000 rows per forward request
// by default); override with WithPageSize or observe progress with
// WithPageCallback.
func (c *Client) ListNormalizedInterfaces(ctx context.Context, adom string, opts ...ListOption) ([]NormalizedInterface, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/dynamic/interface", adom)
	items, err := getPaged[apiNormalizedInterface](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	result := make([]NormalizedInterface, len(items))
	for i, it := range items {
		result[i] = NormalizedInterface{
			Name:           it.Name,
			SingleIntf:     it.SingleIntf != 0,
			ZoneOnly:       it.ZoneOnly != 0,
			Wildcard:       it.Wildcard != 0,
			DefaultMapping: it.DefaultMapping != 0,
			Color:          it.Color,
			Mappings:       flattenDynamicMappings(it.DynamicMapping),
		}
	}
	return result, nil
}

// flattenDynamicMappings fans out each raw dynamic_mapping entry into
// one NormalizedInterfaceMapping per _scope element. Entries with an
// empty _scope (no device binding) are skipped.
func flattenDynamicMappings(raw []apiDynamicIntfMapping) []NormalizedInterfaceMapping {
	if len(raw) == 0 {
		return nil
	}
	var out []NormalizedInterfaceMapping
	for _, m := range raw {
		intrazoneDeny := toString(m.IntrazoneDeny) == "1"
		if len(m.Scope) == 0 {
			continue
		}
		for _, s := range m.Scope {
			out = append(out, NormalizedInterfaceMapping{
				Device:        s.Name,
				VDOM:          s.VDOM,
				LocalIntf:     append([]string(nil), m.LocalIntf...),
				IntrazoneDeny: intrazoneDeny,
			})
		}
	}
	return out
}
