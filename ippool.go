package fortimgr

import (
	"context"
	"fmt"
)

type apiIPPool struct {
	Name    string `json:"name"`
	Type    any    `json:"type"`
	StartIP string `json:"startip"`
	EndIP   string `json:"endip"`
	Comment string `json:"comments"`
}

// ListIPPools retrieves IP pools from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListIPPools(ctx context.Context, adom string, opts ...ListOption) ([]IPPool, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/ippool", adom)
	items, err := getPaged[apiIPPool](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	pools := make([]IPPool, len(items))
	for i, p := range items {
		pools[i] = IPPool{
			Name:    p.Name,
			Type:    mapEnum(toString(p.Type), ippoolTypes),
			StartIP: p.StartIP,
			EndIP:   p.EndIP,
			Comment: p.Comment,
		}
	}

	return pools, nil
}
