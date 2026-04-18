package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
)

type apiADOM struct {
	Name  string `json:"name"`
	OID   int    `json:"oid"`
	Desc  string `json:"desc"`
	State any    `json:"state"`
	Mode  any    `json:"mode"`
	OSVer any    `json:"os_ver"`
	MrNum any    `json:"mr"`
}

// apiSessionScope is the subset of /gui/sys/config used to determine which
// ADOMs the current session can access.
type apiSessionScope struct {
	Adom struct {
		Name string `json:"name"`
	} `json:"adom"`
	Adoms []string `json:"adoms"`
}

// ListADOMs retrieves Administrative Domains from FortiManager.
// ADOMs partition managed devices, policies, and objects into isolated scopes.
//
// By default, only ADOMs accessible to the current session are returned —
// this matches what the FMG GUI shows and excludes factory-preset ADOMs
// the logged-in admin has no scope for. Pass all=true to return every ADOM
// on the system (global /dvmdb/adom view, requires superadmin).
func (c *Client) ListADOMs(ctx context.Context, all ...bool) ([]ADOM, error) {
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

	if len(all) > 0 && all[0] {
		return adoms, nil
	}

	allowed, err := c.sessionADOMScope(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]ADOM, 0, len(allowed))
	for _, a := range adoms {
		if _, ok := allowed[a.Name]; ok {
			filtered = append(filtered, a)
		}
	}
	return filtered, nil
}

// sessionADOMScope returns the set of ADOM names the current session can access,
// read from /gui/sys/config (the same endpoint used by SystemStatus). The set
// contains the current ADOM (data.adom.name) plus any additional accessible
// ADOMs listed in data.adoms.
func (c *Client) sessionADOMScope(ctx context.Context) (map[string]struct{}, error) {
	data, err := c.proxy(ctx, "/gui/sys/config", "get")
	if err != nil {
		return nil, err
	}

	var scope apiSessionScope
	if err := json.Unmarshal(data, &scope); err != nil {
		return nil, fmt.Errorf("fortimgr: parse session scope: %w", err)
	}

	names := make(map[string]struct{}, len(scope.Adoms)+1)
	if scope.Adom.Name != "" {
		names[scope.Adom.Name] = struct{}{}
	}
	for _, n := range scope.Adoms {
		names[n] = struct{}{}
	}
	return names, nil
}

// adomOID resolves an ADOM name to the FlatUI OID used by some /gui/adoms
// endpoints. It intentionally stays internal because OIDs are not stable SDK
// identity; public methods should continue accepting ADOM names.
func (c *Client) adomOID(ctx context.Context, adom string) (int, error) {
	items, err := get[apiADOM](ctx, c, "/dvmdb/adom")
	if err != nil {
		return 0, err
	}
	for _, a := range items {
		if a.Name == adom && a.OID > 0 {
			return a.OID, nil
		}
	}
	return 0, fmt.Errorf("fortimgr: ADOM %q OID not found", adom)
}
