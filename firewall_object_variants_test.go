package fortimgr

import (
	"context"
	"testing"
)

func TestFirewallObjectVariants(t *testing.T) {
	client := newTestClient(t, map[string]string{
		"/pm/config/adom/root/obj/firewall/address6": `[
			{"name":"addr6","type":"ipmask","ip6":"2001:db8::/64","comment":"ipv6","color":3,"associated-interface":"wan1"}
		]`,
		"/pm/config/adom/root/obj/firewall/addrgrp6": `[
			{"name":"grp6","member":["addr6"],"comment":"group","color":4}
		]`,
		"/pm/config/adom/root/obj/firewall/vipgrp": `[
			{"name":"vipgrp","member":["vip1"],"comment":"vip group","color":5}
		]`,
		"/pm/config/adom/root/obj/firewall/vip6": `[
			{"name":"vip6","extip":"2001:db8::10","mappedip":["fd00::10-fd00::10"],"extintf":"wan1","portforward":1,"protocol":1,"extport":"443","mappedport":"8443","comment":"vip6","color":6}
		]`,
		"/pm/config/adom/root/obj/firewall/vipgrp6": `[
			{"name":"vipgrp6","member":["vip6"],"comment":"vip6 group","color":7}
		]`,
		"/pm/config/adom/root/obj/firewall/ippool6": `[
			{"name":"pool6","type":0,"startip":"2001:db8::1","endip":"2001:db8::ffff","comments":"pool6"}
		]`,
		"/pm/config/adom/root/obj/firewall/ippool_grp": `[
			{"name":"poolgrp","member":["pool1"],"comments":"pool group"}
		]`,
		"/pm/config/adom/root/obj/firewall/internet-service-custom": `[
			{"name":"custom-is","id":1001,"comment":"custom","entry":["tcp/443"]}
		]`,
		"/pm/config/adom/root/obj/firewall/internet-service-custom-group": `[
			{"name":"custom-is-grp","id":2001,"member":["custom-is"],"comment":"custom group"}
		]`,
		"/pm/config/adom/root/obj/firewall/internet-service-group": `[
			{"name":"is-grp","id":3001,"member":["Microsoft"],"comment":"group"}
		]`,
		"/pm/config/adom/root/obj/firewall/internet-service-name": `[
			{"name":"Microsoft","id":327786}
		]`,
		"/pm/config/adom/root/obj/_fdsdb/internet-service": `[
			{"name":"Microsoft","id":327786,"category":"Cloud","protocol":"tcp"}
		]`,
		"/pm/config/adom/root/obj/firewall/schedule/group": `[
			{"name":"schedgrp","member":["always"],"comment":"schedule group","color":2}
		]`,
	})

	addrs6, err := client.ListAddresses6(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs6) != 1 || addrs6[0].IP6 != "2001:db8::/64" || addrs6[0].AssocIntf != "wan1" {
		t.Errorf("addrs6 = %+v", addrs6)
	}

	addrgrps6, err := client.ListAddressGroups6(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrgrps6) != 1 || addrgrps6[0].Members[0] != "addr6" {
		t.Errorf("addrgrps6 = %+v", addrgrps6)
	}

	vipgrps, err := client.ListVirtualIPGroups(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(vipgrps) != 1 || vipgrps[0].Members[0] != "vip1" {
		t.Errorf("vipgrps = %+v", vipgrps)
	}

	vips6, err := client.ListVirtualIPs6(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(vips6) != 1 || vips6[0].PortForward != "enable" || vips6[0].Protocol != "tcp" {
		t.Errorf("vips6 = %+v", vips6)
	}

	vipgrps6, err := client.ListVirtualIPGroups6(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(vipgrps6) != 1 || vipgrps6[0].Members[0] != "vip6" {
		t.Errorf("vipgrps6 = %+v", vipgrps6)
	}

	pools6, err := client.ListIPPools6(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(pools6) != 1 || pools6[0].Type != "overload" || pools6[0].StartIP != "2001:db8::1" {
		t.Errorf("pools6 = %+v", pools6)
	}

	poolgrps, err := client.ListIPPoolGroups(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(poolgrps) != 1 || poolgrps[0].Members[0] != "pool1" {
		t.Errorf("poolgrps = %+v", poolgrps)
	}

	customIS, err := client.ListInternetServiceCustom(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(customIS) != 1 || customIS[0].ID != 1001 || customIS[0].Entry[0] != "tcp/443" {
		t.Errorf("customIS = %+v", customIS)
	}

	customISGroups, err := client.ListInternetServiceCustomGroups(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(customISGroups) != 1 || customISGroups[0].Members[0] != "custom-is" {
		t.Errorf("customISGroups = %+v", customISGroups)
	}

	isGroups, err := client.ListInternetServiceGroups(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(isGroups) != 1 || isGroups[0].ID != 3001 {
		t.Errorf("isGroups = %+v", isGroups)
	}

	isNames, err := client.ListInternetServiceNames(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(isNames) != 1 || isNames[0].ID != 327786 {
		t.Errorf("isNames = %+v", isNames)
	}

	fdsdb, err := client.ListFDSDBInternetServices(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(fdsdb) != 1 || fdsdb[0].Category != "Cloud" || fdsdb[0].Protocol != "tcp" {
		t.Errorf("fdsdb = %+v", fdsdb)
	}

	scheduleGroups, err := client.ListScheduleGroups(context.Background(), "root")
	if err != nil {
		t.Fatal(err)
	}
	if len(scheduleGroups) != 1 || scheduleGroups[0].Members[0] != "always" {
		t.Errorf("scheduleGroups = %+v", scheduleGroups)
	}
}
