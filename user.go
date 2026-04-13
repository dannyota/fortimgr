package fortimgr

import (
	"context"
	"fmt"
)

type apiUser struct {
	Name    string `json:"name"`
	Status  any    `json:"status"`
	Type    any    `json:"type"`
	EmailTo string `json:"email-to"`
}

type apiUserGroup struct {
	Name      string `json:"name"`
	Member    any    `json:"member"`
	GroupType any    `json:"group-type"`
	Comment   string `json:"comment"`
}

type apiLDAPServer struct {
	Name   string `json:"name"`
	Server string `json:"server"`
	Port   int    `json:"port"`
	DN     string `json:"dn"`
	Type   any    `json:"type"`
	Secure any    `json:"secure"`
}

type apiRADIUSServer struct {
	Name     string `json:"name"`
	Server   string `json:"server"`
	AuthType any    `json:"auth-type"`
	NASIP    string `json:"nas-ip"`
}

// ListUsers retrieves local user objects from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListUsers(ctx context.Context, adom string, opts ...ListOption) ([]User, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/user/local", adom)
	items, err := getPaged[apiUser](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	users := make([]User, len(items))
	for i, u := range items {
		users[i] = User{
			Name:   u.Name,
			Status: mapEnum(toString(u.Status), enableDisable),
			Type:   mapEnum(toString(u.Type), userTypes),
			Email:  u.EmailTo,
		}
	}

	return users, nil
}

// ListUserGroups retrieves user groups from an ADOM.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListUserGroups(ctx context.Context, adom string, opts ...ListOption) ([]UserGroup, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/user/group", adom)
	items, err := getPaged[apiUserGroup](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	groups := make([]UserGroup, len(items))
	for i, g := range items {
		groups[i] = UserGroup{
			Name:    g.Name,
			Members: toStringSlice(g.Member),
			Type:    mapEnum(toString(g.GroupType), userGroupTypes),
			Comment: g.Comment,
		}
	}

	return groups, nil
}

// ListLDAPServers retrieves LDAP server configurations from an ADOM.
// Credentials are intentionally excluded from the response.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListLDAPServers(ctx context.Context, adom string, opts ...ListOption) ([]LDAPServer, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/user/ldap", adom)
	items, err := getPaged[apiLDAPServer](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	servers := make([]LDAPServer, len(items))
	for i, s := range items {
		servers[i] = LDAPServer{
			Name:   s.Name,
			Server: s.Server,
			Port:   s.Port,
			DN:     s.DN,
			Type:   mapEnum(toString(s.Type), ldapTypes),
			Secure: mapEnum(toString(s.Secure), ldapSecure),
		}
	}

	return servers, nil
}

// ListRADIUSServers retrieves RADIUS server configurations from an ADOM.
// Credentials are intentionally excluded from the response.
// Pagination is applied transparently; see WithPageSize / WithPageCallback.
func (c *Client) ListRADIUSServers(ctx context.Context, adom string, opts ...ListOption) ([]RADIUSServer, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/user/radius", adom)
	items, err := getPaged[apiRADIUSServer](ctx, c, apiURL, nil, buildListConfig(opts))
	if err != nil {
		return nil, err
	}

	servers := make([]RADIUSServer, len(items))
	for i, s := range items {
		servers[i] = RADIUSServer{
			Name:     s.Name,
			Server:   s.Server,
			AuthType: mapEnum(toString(s.AuthType), radiusAuthTypes),
			NASIP:    s.NASIP,
		}
	}

	return servers, nil
}
