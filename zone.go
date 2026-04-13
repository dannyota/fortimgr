package fortimgr

import (
	"context"
	"fmt"
)

type apiZone struct {
	Name        string `json:"name"`
	Interface   any    `json:"interface"`
	Intrazone   any    `json:"intrazone"`
	Description string `json:"description"`
}

// ListZones retrieves system zones from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListZones(ctx context.Context, adom string, opts ...ListOption) ([]Zone, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/system/zone", adom)
	items, err := getPaged[apiZone](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	zones := make([]Zone, len(items))
	for i, z := range items {
		zones[i] = Zone{
			Name:        z.Name,
			Interfaces:  toStringSlice(z.Interface),
			Intrazone:   mapEnum(toString(z.Intrazone), intrazoneTraffic),
			Description: z.Description,
		}
	}

	return zones, nil
}
