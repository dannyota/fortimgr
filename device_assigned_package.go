package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
)

type apiAssignedPackageRef struct {
	Name   string `json:"name"`
	OID    int    `json:"oid"`
	Flags  int    `json:"flags"`
	Status int    `json:"status"`
}

type apiDeviceAssignedPackage struct {
	DeviceOID    int                   `json:"deviceOid"`
	VDOMOID      int                   `json:"vdomOid"`
	Pkg          apiAssignedPackageRef `json:"pkg"`
	FAPProfile   apiAssignedPackageRef `json:"fap_prof"`
	FExtProfile  apiAssignedPackageRef `json:"fext_prof"`
	ProfileDirty bool                  `json:"profileDirty"`
}

// ListDeviceAssignedPackages retrieves the policy/profile package assignments
// shown by FortiManager's Device Manager for every device/VDOM in an ADOM.
//
// This FlatUI endpoint returns an aggregate map rather than a paginated list,
// so there are no pagination options to expose.
func (c *Client) ListDeviceAssignedPackages(ctx context.Context, adom string) ([]DeviceAssignedPackage, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	oid, err := c.adomOID(ctx, adom)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("/gui/adoms/%d/devices/assignedpkgs", oid)
	data, err := c.proxy(ctx, apiURL, "get")
	if err != nil {
		return nil, err
	}

	var raw map[string]apiDeviceAssignedPackage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal assigned packages: %w", err)
	}

	items := make([]DeviceAssignedPackage, 0, len(raw))
	for _, p := range raw {
		items = append(items, DeviceAssignedPackage{
			DeviceOID:    p.DeviceOID,
			VDOMOID:      p.VDOMOID,
			Package:      assignedPackageRef(p.Pkg),
			FAPProfile:   assignedPackageRef(p.FAPProfile),
			FExtProfile:  assignedPackageRef(p.FExtProfile),
			ProfileDirty: p.ProfileDirty,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].DeviceOID != items[j].DeviceOID {
			return items[i].DeviceOID < items[j].DeviceOID
		}
		if items[i].VDOMOID != items[j].VDOMOID {
			return items[i].VDOMOID < items[j].VDOMOID
		}
		return items[i].Package.Name < items[j].Package.Name
	})
	return items, nil
}

func assignedPackageRef(p apiAssignedPackageRef) AssignedPackageRef {
	return AssignedPackageRef(p)
}
