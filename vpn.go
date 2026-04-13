package fortimgr

import (
	"context"
	"fmt"
)

type apiIPSecPhase1 struct {
	Name      string `json:"name"`
	Interface any    `json:"interface"`
	RemoteGW  string `json:"remote-gw"`
	Proposal  any    `json:"proposal"`
	DHGroup   any    `json:"dhgrp"`
	Mode      any    `json:"mode"`
	Type      any    `json:"type"`
	Comments  string `json:"comments"`
}

type apiIPSecPhase2 struct {
	Name       string `json:"name"`
	Phase1Name any    `json:"phase1name"`
	Proposal   any    `json:"proposal"`
	SrcSubnet  any    `json:"src-subnet"`
	DstSubnet  any    `json:"dst-subnet"`
	Comments   string `json:"comments"`
}

// ListIPSecPhase1 retrieves IPSec Phase 1 interface configurations from an ADOM.
func (c *Client) ListIPSecPhase1(ctx context.Context, adom string) ([]IPSecPhase1, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/vpn/ipsec/phase1-interface", adom)
	items, err := get[apiIPSecPhase1](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	tunnels := make([]IPSecPhase1, len(items))
	for i, p := range items {
		tunnels[i] = IPSecPhase1{
			Name:      p.Name,
			Interface: toString(p.Interface),
			RemoteGW:  p.RemoteGW,
			Proposal:  toString(p.Proposal),
			DHGroup:   toString(p.DHGroup),
			Mode:      mapEnum(toString(p.Mode), ipsecModes),
			Type:      mapEnum(toString(p.Type), ipsecTypes),
			Comments:  p.Comments,
		}
	}

	return tunnels, nil
}

// ListIPSecPhase2 retrieves IPSec Phase 2 interface configurations from an ADOM.
func (c *Client) ListIPSecPhase2(ctx context.Context, adom string) ([]IPSecPhase2, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/vpn/ipsec/phase2-interface", adom)
	items, err := get[apiIPSecPhase2](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	tunnels := make([]IPSecPhase2, len(items))
	for i, p := range items {
		tunnels[i] = IPSecPhase2{
			Name:       p.Name,
			Phase1Name: toString(p.Phase1Name),
			Proposal:   toString(p.Proposal),
			SrcSubnet:  formatSubnet(p.SrcSubnet),
			DstSubnet:  formatSubnet(p.DstSubnet),
			Comments:   p.Comments,
		}
	}

	return tunnels, nil
}
