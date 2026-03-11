package fortimgr

import "context"

type apiADOM struct {
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	State any    `json:"state"`
	Mode  any    `json:"mode"`
	OSVer any    `json:"os_ver"`
	MrNum any    `json:"mr"`
}

// ListADOMs retrieves all Administrative Domains from FortiManager.
// ADOMs partition managed devices, policies, and objects into isolated scopes.
// Use "root" for the default ADOM in single-tenant deployments.
func (c *Client) ListADOMs(ctx context.Context) ([]ADOM, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	items, err := get[apiADOM](ctx, c, "/dvmdb/adom")
	if err != nil {
		return nil, err
	}

	adoms := make([]ADOM, len(items))
	for i, a := range items {
		// OS version: e.g. os_ver=7 mr=2 → "7.2"
		osVer := toString(a.OSVer)
		mr := toString(a.MrNum)
		ver := ""
		if osVer != "0" && osVer != "" {
			ver = osVer
			if mr != "" && mr != "0" {
				ver += "." + mr
			}
		}

		adoms[i] = ADOM{
			Name:  a.Name,
			Desc:  a.Desc,
			State: mapEnum(toString(a.State), adomStates),
			Mode:  mapEnum(toString(a.Mode), adomModes),
			OSVer: ver,
		}
	}

	return adoms, nil
}
