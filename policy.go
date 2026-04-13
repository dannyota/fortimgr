package fortimgr

import (
	"context"
	"fmt"
)

type apiPolicyPackage struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Scope []struct {
		Name string `json:"name"`
	} `json:"scope member"`
}

type apiPolicy struct {
	PolicyID   int    `json:"policyid"`
	Name       string `json:"name"`
	SrcIntf    any    `json:"srcintf"`
	DstIntf    any    `json:"dstintf"`
	SrcAddr    any    `json:"srcaddr"`
	DstAddr    any    `json:"dstaddr"`
	Service    any    `json:"service"`
	Action     any    `json:"action"`
	Schedule   any    `json:"schedule"`
	NAT        any    `json:"nat"`
	Status     any    `json:"status"`
	LogTraffic any    `json:"logtraffic"`
	Comments   string `json:"comments"`
}

// ListPolicyPackages retrieves all policy packages from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListPolicyPackages(ctx context.Context, adom string, opts ...ListOption) ([]PolicyPackage, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/pkg/adom/%s", adom)
	items, err := getPaged[apiPolicyPackage](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	packages := make([]PolicyPackage, 0, len(items))
	for _, p := range items {
		if p.Type != "pkg" {
			continue
		}
		scope := make([]string, len(p.Scope))
		for j, s := range p.Scope {
			scope[j] = s.Name
		}
		packages = append(packages, PolicyPackage{
			Name:  p.Name,
			Type:  p.Type,
			ADOM:  adom,
			Scope: scope,
		})
	}

	return packages, nil
}

// ListPolicies retrieves firewall policies from a policy package.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListPolicies(ctx context.Context, adom, pkg string, opts ...ListOption) ([]Policy, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) || !validName(pkg) {
		return nil, fmt.Errorf("%w: adom=%q pkg=%q", ErrInvalidName, adom, pkg)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/pkg/%s/firewall/policy", adom, pkg)
	items, err := getPaged[apiPolicy](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	policies := make([]Policy, len(items))
	for i, p := range items {
		policies[i] = Policy{
			PolicyID:   p.PolicyID,
			Name:       p.Name,
			SrcIntf:    toStringSlice(p.SrcIntf),
			DstIntf:    toStringSlice(p.DstIntf),
			SrcAddr:    toStringSlice(p.SrcAddr),
			DstAddr:    toStringSlice(p.DstAddr),
			Service:    toStringSlice(p.Service),
			Action:     mapEnum(toString(p.Action), policyActions),
			Schedule:   toString(p.Schedule),
			NAT:        mapEnum(toString(p.NAT), enableDisable),
			Status:     mapEnum(toString(p.Status), enableDisable),
			LogTraffic: mapEnum(toString(p.LogTraffic), logTrafficModes),
			Comments:   p.Comments,
		}
	}

	return policies, nil
}
