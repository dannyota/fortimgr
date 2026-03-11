package fortimgr

import (
	"context"
	"fmt"
)

type apiVDOM struct {
	Name   string `json:"name"`
	Status any    `json:"status"`
	OpMode any    `json:"opmode"`
}

// ListVDOMs retrieves Virtual Domains from a FortiGate device.
func (c *Client) ListVDOMs(ctx context.Context, device string) ([]VDOM, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(device) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, device)
	}

	apiURL := fmt.Sprintf("/dvmdb/device/%s/vdom", device)
	items, err := get[apiVDOM](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	vdoms := make([]VDOM, len(items))
	for i, v := range items {
		vdoms[i] = VDOM{
			Name:   v.Name,
			Status: mapEnum(toString(v.Status), enableDisable),
			OpMode: mapEnum(toString(v.OpMode), vdomOpModes),
		}
	}

	return vdoms, nil
}
