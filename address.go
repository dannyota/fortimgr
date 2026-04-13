package fortimgr

import (
	"context"
	"fmt"
)

type apiAddress struct {
	Name      string `json:"name"`
	Type      any    `json:"type"`
	Subnet    any    `json:"subnet"`
	StartIP   string `json:"start-ip"`
	EndIP     string `json:"end-ip"`
	FQDN      string `json:"fqdn"`
	Country   any    `json:"country"`
	Wildcard  any    `json:"wildcard"`
	Comment   string `json:"comment"`
	Color     int    `json:"color"`
	AssocIntf any    `json:"associated-interface"`
}

type apiAddressGroup struct {
	Name    string `json:"name"`
	Member  any    `json:"member"`
	Comment string `json:"comment"`
	Color   int    `json:"color"`
}

// ListAddresses retrieves firewall address objects from an ADOM.
//
// Pagination: this method transparently fetches every page from the
// FortiManager API. The default page size is 1000 rows per forward
// request; override with WithPageSize. Progress can be observed via
// WithPageCallback. See the ListOption godoc for details.
func (c *Client) ListAddresses(ctx context.Context, adom string, opts ...ListOption) ([]Address, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/address", adom)
	items, err := getPaged[apiAddress](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	addresses := make([]Address, len(items))
	for i, a := range items {
		addresses[i] = Address{
			Name:      a.Name,
			Type:      mapEnum(toString(a.Type), addressTypes),
			Subnet:    formatSubnet(a.Subnet),
			StartIP:   a.StartIP,
			EndIP:     a.EndIP,
			FQDN:      a.FQDN,
			Country:   toString(a.Country),
			Wildcard:  formatSubnet(a.Wildcard),
			Comment:   a.Comment,
			Color:     a.Color,
			AssocIntf: toString(a.AssocIntf),
		}
	}

	return addresses, nil
}

// ListAddressGroups retrieves firewall address groups from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListAddressGroups(ctx context.Context, adom string, opts ...ListOption) ([]AddressGroup, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/addrgrp", adom)
	items, err := getPaged[apiAddressGroup](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	groups := make([]AddressGroup, len(items))
	for i, g := range items {
		groups[i] = AddressGroup{
			Name:    g.Name,
			Members: toStringSlice(g.Member),
			Comment: g.Comment,
			Color:   g.Color,
		}
	}

	return groups, nil
}
