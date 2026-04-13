//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"danny.vn/fortimgr"
)

type config struct {
	Address            string `json:"address"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	ADOM               string `json:"adom"`
	InsecureTLS        bool   `json:"insecure_tls"`
	X509NegativeSerial bool   `json:"x509_negative_serial"`
}

func loadConfig() config {
	var cfg config

	// Load from .fortimgr.json if it exists.
	if data, err := os.ReadFile(".fortimgr.json"); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Fatalf("invalid .fortimgr.json: %v", err)
		}
	}

	// Env vars override file values.
	if v := os.Getenv("FORTIMGR_ADDRESS"); v != "" {
		cfg.Address = v
	}
	if v := os.Getenv("FORTIMGR_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("FORTIMGR_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("FORTIMGR_ADOM"); v != "" {
		cfg.ADOM = v
	}

	if cfg.Address == "" || cfg.Username == "" || cfg.Password == "" {
		fmt.Fprintln(os.Stderr, "Create .fortimgr.json or set FORTIMGR_ADDRESS, FORTIMGR_USERNAME, FORTIMGR_PASSWORD")
		os.Exit(1)
	}

	return cfg
}

// writeSample writes a JSON sample file to samples/<name>.json.
func writeSample(name string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: marshal %s sample: %v\n", name, err)
		return
	}
	os.MkdirAll("samples", 0o755)
	path := fmt.Sprintf("samples/%s.json", name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "warning: write %s: %v\n", path, err)
		return
	}
	fmt.Printf("  → %s (%d bytes)\n", path, len(data))
}

func main() {
	cfg := loadConfig()

	var opts []fortimgr.ClientOption
	opts = append(opts, fortimgr.WithCredentials(cfg.Username, cfg.Password))
	if cfg.InsecureTLS {
		opts = append(opts, fortimgr.WithInsecureTLS())
	}
	if cfg.X509NegativeSerial {
		opts = append(opts, fortimgr.WithX509NegativeSerial())
	}

	ctx := context.Background()

	client, err := fortimgr.NewClient(cfg.Address, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	fmt.Printf("Connecting to %s...\n", cfg.Address)
	if err := client.Login(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Login OK")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)

	// ADOMs — default call returns only ADOMs accessible to the current session.
	adoms, err := client.ListADOMs(ctx)
	if err != nil {
		log.Fatalf("ListADOMs: %v", err)
	}
	// Global view — every ADOM on the box, including factory presets the session
	// admin has no scope for. Useful as a diagnostic against the filtered list.
	allADOMs, err := client.ListADOMs(ctx, true)
	if err != nil {
		log.Fatalf("ListADOMs(all=true): %v", err)
	}
	fmt.Printf("ADOMs: %d accessible / %d total\n", len(adoms), len(allADOMs))
	fmt.Fprintf(w, "ADOM\tSTATE\tMODE\tOS VERSION\tDESCRIPTION\n")
	for _, a := range adoms {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			a.Name, a.State, a.Mode, a.OSVer, a.Desc)
	}
	w.Flush()
	fmt.Println()

	// Use configured ADOM, or auto-detect: prefer "root", then first enabled.
	if cfg.ADOM == "" {
		for _, a := range adoms {
			if a.Name == "root" && a.State == "enabled" {
				cfg.ADOM = "root"
				break
			}
		}
		if cfg.ADOM == "" {
			for _, a := range adoms {
				if a.State == "enabled" {
					cfg.ADOM = a.Name
					break
				}
			}
		}
		if cfg.ADOM == "" && len(adoms) > 0 {
			cfg.ADOM = adoms[0].Name
		}
		if cfg.ADOM == "" {
			log.Fatal("no ADOMs found")
		}
		fmt.Printf("Auto-selected ADOM: %s\n\n", cfg.ADOM)
	} else {
		fmt.Printf("Using ADOM: %s\n\n", cfg.ADOM)
	}

	// Devices.
	devices, err := client.ListDevices(ctx, cfg.ADOM)
	if err != nil {
		log.Fatalf("ListDevices: %v", err)
	}
	fmt.Fprintf(w, "DEVICE\tSERIAL\tPLATFORM\tFIRMWARE\tHA\tSTATUS\tIP\n")
	for _, d := range devices {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			d.Name, d.SerialNumber, d.Platform, d.Firmware, d.HAMode, d.Status, d.IPAddress)
	}
	w.Flush()
	fmt.Println()

	// Policy packages + policies per package.
	pkgs, err := client.ListPolicyPackages(ctx, cfg.ADOM)
	if err != nil {
		log.Fatalf("ListPolicyPackages: %v", err)
	}
	fmt.Printf("Policy Packages: %d\n", len(pkgs))

	var allPolicies []fortimgr.Policy
	for _, p := range pkgs {
		policies, err := client.ListPolicies(ctx, cfg.ADOM, p.Name)
		if err != nil {
			fmt.Printf("  %s: error: %v\n", p.Name, err)
			continue
		}
		fmt.Printf("  %s: %d policies (scope: %v)\n", p.Name, len(policies), p.Scope)
		allPolicies = append(allPolicies, policies...)
	}
	fmt.Println()

	// Collect all resources.
	type resource struct {
		name  string
		count int
		err   error
	}
	var resources []resource

	addrs, err := client.ListAddresses(ctx, cfg.ADOM)
	resources = append(resources, resource{"Addresses", len(addrs), err})

	groups, err := client.ListAddressGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Address Groups", len(groups), err})

	services, err := client.ListServices(ctx, cfg.ADOM)
	resources = append(resources, resource{"Services", len(services), err})

	svcGroups, err := client.ListServiceGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Service Groups", len(svcGroups), err})

	recurring, err := client.ListSchedulesRecurring(ctx, cfg.ADOM)
	resources = append(resources, resource{"Schedules (recurring)", len(recurring), err})

	onetime, err := client.ListSchedulesOnetime(ctx, cfg.ADOM)
	resources = append(resources, resource{"Schedules (onetime)", len(onetime), err})

	virtualIPs, err := client.ListVirtualIPs(ctx, cfg.ADOM)
	resources = append(resources, resource{"Virtual IPs", len(virtualIPs), err})

	pools, err := client.ListIPPools(ctx, cfg.ADOM)
	resources = append(resources, resource{"IP Pools", len(pools), err})

	zones, err := client.ListZones(ctx, cfg.ADOM)
	resources = append(resources, resource{"Zones", len(zones), err})

	avProfiles, err := client.ListAntivirusProfiles(ctx, cfg.ADOM)
	resources = append(resources, resource{"Antivirus Profiles", len(avProfiles), err})

	ipsSensors, err := client.ListIPSSensors(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPS Sensors", len(ipsSensors), err})

	wfProfiles, err := client.ListWebFilterProfiles(ctx, cfg.ADOM)
	resources = append(resources, resource{"Web Filter Profiles", len(wfProfiles), err})

	appProfiles, err := client.ListAppControlProfiles(ctx, cfg.ADOM)
	resources = append(resources, resource{"App Control Profiles", len(appProfiles), err})

	sslProfiles, err := client.ListSSLSSHProfiles(ctx, cfg.ADOM)
	resources = append(resources, resource{"SSL/SSH Profiles", len(sslProfiles), err})

	users, err := client.ListUsers(ctx, cfg.ADOM)
	resources = append(resources, resource{"Users", len(users), err})

	userGroups, err := client.ListUserGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"User Groups", len(userGroups), err})

	ldapServers, err := client.ListLDAPServers(ctx, cfg.ADOM)
	resources = append(resources, resource{"LDAP Servers", len(ldapServers), err})

	radiusServers, err := client.ListRADIUSServers(ctx, cfg.ADOM)
	resources = append(resources, resource{"RADIUS Servers", len(radiusServers), err})

	phase1, err := client.ListIPSecPhase1(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPSec Phase 1", len(phase1), err})

	phase2, err := client.ListIPSecPhase2(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPSec Phase 2", len(phase2), err})

	installStatus, err := client.ListPackageInstallStatus(ctx, cfg.ADOM, "")
	resources = append(resources, resource{"Package Install Status", len(installStatus), err})

	// Device-scoped resources (interfaces, routes). VDOMs are derived from the
	// interface list's "vdom" field — this avoids /dvmdb/device/<dev>/vdom,
	// which is permission-denied for restricted admins.
	vdomSeen := map[string]struct{}{}
	var allVDOMs []fortimgr.VDOM
	var allInterfaces []fortimgr.Interface
	var allRoutes []fortimgr.StaticRoute
	for _, d := range devices {
		// Empty vdom → device-wide /global/system/interface path.
		ifaces, err := client.ListInterfaces(ctx, d.Name, "")
		if err != nil {
			fmt.Printf("  Interfaces(%s): %v\n", d.Name, err)
			continue
		}
		allInterfaces = append(allInterfaces, ifaces...)

		// Derive unique VDOMs for this device from the interface list.
		deviceVDOMs := map[string]struct{}{}
		for _, iface := range ifaces {
			if iface.VDOM == "" {
				continue
			}
			deviceVDOMs[iface.VDOM] = struct{}{}
		}
		for v := range deviceVDOMs {
			key := d.Name + "/" + v
			if _, ok := vdomSeen[key]; ok {
				continue
			}
			vdomSeen[key] = struct{}{}
			allVDOMs = append(allVDOMs, fortimgr.VDOM{Name: v, Status: "enable"})

			routes, err := client.ListStaticRoutes(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  Routes(%s/%s): %v\n", d.Name, v, err)
				continue
			}
			allRoutes = append(allRoutes, routes...)
		}
	}

	// System status.
	sysStatus, sysErr := client.SystemStatus(ctx)

	// Device firmware.
	firmware, fwErr := client.ListDeviceFirmware(ctx)

	// Summary table.
	fmt.Fprintf(w, "RESOURCE\tCOUNT\tSTATUS\n")
	if sysErr != nil {
		fmt.Fprintf(w, "System Status\t0\t%s\n", sysErr)
	} else {
		fmt.Fprintf(w, "System Status\t1\t%s %s (build %d)\n", sysStatus.Hostname, sysStatus.Version, sysStatus.Build)
	}
	fmt.Fprintf(w, "ADOMs (accessible)\t%d\tOK\n", len(adoms))
	fmt.Fprintf(w, "ADOMs (total)\t%d\tOK\n", len(allADOMs))
	fmt.Fprintf(w, "Devices\t%d\tOK\n", len(devices))
	fmt.Fprintf(w, "VDOMs\t%d\tOK\n", len(allVDOMs))
	fmt.Fprintf(w, "Interfaces\t%d\tOK\n", len(allInterfaces))
	fmt.Fprintf(w, "Static Routes\t%d\tOK\n", len(allRoutes))
	if fwErr != nil {
		fmt.Fprintf(w, "Device Firmware\t0\t%s\n", fwErr)
	} else {
		fmt.Fprintf(w, "Device Firmware\t%d\tOK\n", len(firmware))
	}
	for _, r := range resources {
		status := "OK"
		if r.err != nil {
			status = r.err.Error()
		}
		fmt.Fprintf(w, "%s\t%d\t%s\n", r.name, r.count, status)
	}
	w.Flush()
	fmt.Println()

	// Write samples for field inspection.
	fmt.Println("Writing samples/...")
	writeSample("adoms", adoms)
	writeSample("adoms_all", allADOMs)
	writeSample("devices", devices)
	writeSample("policy_packages", pkgs)
	writeSample("policies", firstN(allPolicies, 5))
	writeSample("addresses", firstN(addrs, 10))
	writeSample("address_groups", firstN(groups, 5))
	writeSample("services", firstN(services, 10))
	writeSample("service_groups", firstN(svcGroups, 5))
	writeSample("schedules_recurring", firstN(recurring, 5))
	writeSample("schedules_onetime", firstN(onetime, 5))
	writeSample("virtual_ips", firstN(virtualIPs, 5))
	writeSample("ip_pools", firstN(pools, 5))
	writeSample("zones", firstN(zones, 5))
	writeSample("antivirus_profiles", firstN(avProfiles, 5))
	writeSample("ips_sensors", firstN(ipsSensors, 5))
	writeSample("webfilter_profiles", firstN(wfProfiles, 5))
	writeSample("appcontrol_profiles", firstN(appProfiles, 5))
	writeSample("sslssh_profiles", firstN(sslProfiles, 5))
	writeSample("users", firstN(users, 5))
	writeSample("user_groups", firstN(userGroups, 5))
	writeSample("ldap_servers", firstN(ldapServers, 5))
	writeSample("radius_servers", firstN(radiusServers, 5))
	writeSample("ipsec_phase1", firstN(phase1, 5))
	writeSample("ipsec_phase2", firstN(phase2, 5))
	writeSample("package_install_status", installStatus)
	writeSample("vdoms", firstN(allVDOMs, 5))
	writeSample("interfaces", firstN(allInterfaces, 10))
	writeSample("static_routes", firstN(allRoutes, 10))
	if sysStatus != nil {
		writeSample("system_status", sysStatus)
	}
	writeSample("device_firmware", firstN(firmware, 10))
}

// firstN returns the first n items from a slice (or all if fewer).
func firstN[T any](s []T, n int) []T {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
