package fortimgr

import (
	"context"
	"testing"
)

func TestListUsers(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListUsers(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/user/local": `[
				{
					"name": "jdoe",
					"status": 1,
					"type": 1,
					"email-to": "jdoe@example.com"
				},
				{
					"name": "ldap-user",
					"status": "enable",
					"type": 4,
					"email-to": ""
				},
				{
					"name": "disabled-user",
					"status": 0,
					"type": 2,
					"email-to": ""
				}
			]`,
		})

		users, err := client.ListUsers(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(users) != 3 {
			t.Fatalf("len = %d, want 3", len(users))
		}

		u := users[0]
		if u.Name != "jdoe" {
			t.Errorf("Name = %q", u.Name)
		}
		if u.Status != "enable" {
			t.Errorf("Status = %q, want \"enable\"", u.Status)
		}
		if u.Type != "local" {
			t.Errorf("Type = %q, want \"local\"", u.Type)
		}
		if u.Email != "jdoe@example.com" {
			t.Errorf("Email = %q", u.Email)
		}

		if users[1].Type != "ldap" {
			t.Errorf("Type = %q, want \"ldap\"", users[1].Type)
		}
		if users[2].Status != "disable" {
			t.Errorf("Status = %q, want \"disable\"", users[2].Status)
		}
		if users[2].Type != "radius" {
			t.Errorf("Type = %q, want \"radius\"", users[2].Type)
		}
	})
}

func TestListUserGroups(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListUserGroups(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/user/group": `[
				{
					"name": "VPN-Users",
					"member": ["jdoe", "jsmith", "admin"],
					"group-type": 0,
					"comment": "VPN access group"
				},
				{
					"name": "FSSO-Group",
					"member": ["CN=Domain Users"],
					"group-type": 1,
					"comment": ""
				},
				{
					"name": "Guest-Group",
					"member": [],
					"group-type": "guest",
					"comment": "Guest portal"
				}
			]`,
		})

		groups, err := client.ListUserGroups(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(groups) != 3 {
			t.Fatalf("len = %d, want 3", len(groups))
		}

		g := groups[0]
		if g.Name != "VPN-Users" {
			t.Errorf("Name = %q", g.Name)
		}
		if len(g.Members) != 3 {
			t.Errorf("Members = %v", g.Members)
		}
		if g.Type != "firewall" {
			t.Errorf("Type = %q, want \"firewall\"", g.Type)
		}

		if groups[1].Type != "fsso-service" {
			t.Errorf("Type = %q, want \"fsso-service\"", groups[1].Type)
		}
		if groups[2].Type != "guest" {
			t.Errorf("Type = %q, want \"guest\"", groups[2].Type)
		}
		if len(groups[2].Members) != 0 {
			t.Errorf("Members = %v, want empty", groups[2].Members)
		}
	})
}

func TestListLDAPServers(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListLDAPServers(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/user/ldap": `[
				{
					"name": "corp-ldap",
					"server": "ldap.example.com",
					"port": 636,
					"dn": "dc=example,dc=com",
					"type": 2,
					"secure": 2
				},
				{
					"name": "backup-ldap",
					"server": "ldap2.example.com",
					"port": 389,
					"dn": "ou=users,dc=example,dc=com",
					"type": "anonymous",
					"secure": 0
				}
			]`,
		})

		servers, err := client.ListLDAPServers(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(servers) != 2 {
			t.Fatalf("len = %d, want 2", len(servers))
		}

		s := servers[0]
		if s.Name != "corp-ldap" {
			t.Errorf("Name = %q", s.Name)
		}
		if s.Server != "ldap.example.com" {
			t.Errorf("Server = %q", s.Server)
		}
		if s.Port != 636 {
			t.Errorf("Port = %d, want 636", s.Port)
		}
		if s.DN != "dc=example,dc=com" {
			t.Errorf("DN = %q", s.DN)
		}
		if s.Type != "regular" {
			t.Errorf("Type = %q, want \"regular\"", s.Type)
		}
		if s.Secure != "ldaps" {
			t.Errorf("Secure = %q, want \"ldaps\"", s.Secure)
		}

		s2 := servers[1]
		if s2.Type != "anonymous" {
			t.Errorf("Type = %q, want \"anonymous\"", s2.Type)
		}
		if s2.Secure != "disable" {
			t.Errorf("Secure = %q, want \"disable\"", s2.Secure)
		}
	})
}

func TestListRADIUSServers(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		c, _ := NewClient("https://example.com", WithCredentials("u", "p"))
		_, err := c.ListRADIUSServers(context.Background(), "root")
		if err != ErrNotLoggedIn {
			t.Errorf("err = %v, want ErrNotLoggedIn", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		client := newTestClient(t, map[string]string{
			"/pm/config/adom/root/obj/user/radius": `[
				{
					"name": "corp-radius",
					"server": "radius.example.com",
					"auth-type": 0,
					"nas-ip": "10.0.0.1"
				},
				{
					"name": "backup-radius",
					"server": "radius2.example.com",
					"auth-type": 4,
					"nas-ip": ""
				}
			]`,
		})

		servers, err := client.ListRADIUSServers(context.Background(), "root")
		if err != nil {
			t.Fatal(err)
		}
		if len(servers) != 2 {
			t.Fatalf("len = %d, want 2", len(servers))
		}

		if servers[0].AuthType != "auto" {
			t.Errorf("AuthType = %q, want \"auto\"", servers[0].AuthType)
		}
		if servers[0].NASIP != "10.0.0.1" {
			t.Errorf("NASIP = %q", servers[0].NASIP)
		}
		if servers[1].AuthType != "pap" {
			t.Errorf("AuthType = %q, want \"pap\"", servers[1].AuthType)
		}
	})
}
