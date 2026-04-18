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

	fmt.Println("Connecting to FortiManager...")
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

	// Demonstrate WithPageCallback: print progress as each page lands.
	// Default page size (1000) means most ADOMs return everything on
	// page 1; this fires for exactly those deployments where pagination
	// is meaningful.
	addrs, err := client.ListAddresses(ctx, cfg.ADOM,
		fortimgr.WithPageCallback(func(fetched, page int) {
			fmt.Printf("  ListAddresses page %d → %d rows cumulative\n", page, fetched)
		}),
	)
	resources = append(resources, resource{"Addresses", len(addrs), err})

	addrs6, err := client.ListAddresses6(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPv6 Addresses", len(addrs6), err})

	groups, err := client.ListAddressGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Address Groups", len(groups), err})

	groups6, err := client.ListAddressGroups6(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPv6 Address Groups", len(groups6), err})

	services, err := client.ListServices(ctx, cfg.ADOM)
	resources = append(resources, resource{"Services", len(services), err})

	svcGroups, err := client.ListServiceGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Service Groups", len(svcGroups), err})

	recurring, err := client.ListSchedulesRecurring(ctx, cfg.ADOM)
	resources = append(resources, resource{"Schedules (recurring)", len(recurring), err})

	onetime, err := client.ListSchedulesOnetime(ctx, cfg.ADOM)
	resources = append(resources, resource{"Schedules (onetime)", len(onetime), err})

	scheduleGroups, err := client.ListScheduleGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Schedule Groups", len(scheduleGroups), err})

	virtualIPs, err := client.ListVirtualIPs(ctx, cfg.ADOM)
	resources = append(resources, resource{"Virtual IPs", len(virtualIPs), err})

	virtualIPGroups, err := client.ListVirtualIPGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Virtual IP Groups", len(virtualIPGroups), err})

	virtualIPs6, err := client.ListVirtualIPs6(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPv6 Virtual IPs", len(virtualIPs6), err})

	virtualIPGroups6, err := client.ListVirtualIPGroups6(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPv6 Virtual IP Groups", len(virtualIPGroups6), err})

	pools, err := client.ListIPPools(ctx, cfg.ADOM)
	resources = append(resources, resource{"IP Pools", len(pools), err})

	pools6, err := client.ListIPPools6(ctx, cfg.ADOM)
	resources = append(resources, resource{"IPv6 IP Pools", len(pools6), err})

	poolGroups, err := client.ListIPPoolGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"IP Pool Groups", len(poolGroups), err})

	internetServiceCustom, err := client.ListInternetServiceCustom(ctx, cfg.ADOM)
	resources = append(resources, resource{"Internet Service Custom", len(internetServiceCustom), err})

	internetServiceCustomGroups, err := client.ListInternetServiceCustomGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Internet Service Custom Groups", len(internetServiceCustomGroups), err})

	internetServiceGroups, err := client.ListInternetServiceGroups(ctx, cfg.ADOM)
	resources = append(resources, resource{"Internet Service Groups", len(internetServiceGroups), err})

	internetServiceNames, err := client.ListInternetServiceNames(ctx, cfg.ADOM)
	resources = append(resources, resource{"Internet Service Names", len(internetServiceNames), err})

	fdsdbInternetServices, err := client.ListFDSDBInternetServices(ctx, cfg.ADOM)
	resources = append(resources, resource{"FortiGuard Internet Services", len(fdsdbInternetServices), err})

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

	adomRevisions, err := client.ListADOMRevisions(ctx, cfg.ADOM)
	resources = append(resources, resource{"ADOM Revisions", len(adomRevisions), err})

	workflowSessions, err := client.ListWorkflowSessions(ctx, cfg.ADOM)
	resources = append(resources, resource{"Workflow Sessions", len(workflowSessions), err})

	var workflowLogs []fortimgr.WorkflowLog
	if len(workflowSessions) > 0 {
		workflowLogs, err = client.ListWorkflowLogs(ctx, cfg.ADOM, workflowSessions[0].SessionID)
		resources = append(resources, resource{"Workflow Logs", len(workflowLogs), err})
	} else {
		resources = append(resources, resource{"Workflow Logs", 0, nil})
	}

	normIfaces, err := client.ListNormalizedInterfaces(ctx, cfg.ADOM)
	resources = append(resources, resource{"Normalized Interfaces", len(normIfaces), err})

	assignedPkgs, err := client.ListDeviceAssignedPackages(ctx, cfg.ADOM)
	resources = append(resources, resource{"Device Assigned Packages", len(assignedPkgs), err})

	var deviceSummary *fortimgr.DeviceSummary
	var deviceSummaryErr error
	if len(devices) > 0 {
		deviceSummary, deviceSummaryErr = client.DeviceSummary(ctx, cfg.ADOM, devices[0].Name)
		count := 0
		if deviceSummaryErr == nil && deviceSummary != nil {
			count = 1
		}
		resources = append(resources, resource{"Device Summary", count, deviceSummaryErr})
	} else {
		resources = append(resources, resource{"Device Summary", 0, nil})
	}

	// Policy revision counts + per-policy revisions for the first package.
	var revCounts map[int]int
	var policyRevisions []fortimgr.PolicyRevision
	if len(pkgs) > 0 {
		revCounts, err = client.ListPolicyRevisionCounts(ctx, cfg.ADOM, pkgs[0].Name)
		resources = append(resources, resource{"Policy Revision Counts", len(revCounts), err})

		// Pick the first policy with revisions to fetch its history.
		for pid := range revCounts {
			policyRevisions, err = client.ListPolicyRevisions(ctx, cfg.ADOM, pkgs[0].Name, pid)
			resources = append(resources, resource{fmt.Sprintf("Policy Revisions (id=%d)", pid), len(policyRevisions), err})
			break
		}
	}

	// Device-scoped resources (interfaces, routes). VDOM names are derived
	// from the interface list's "vdom" field.
	vdomSeen := map[string]struct{}{}
	var allVDOMNames []string
	var allInterfaces []fortimgr.Interface
	var allRoutes []fortimgr.StaticRoute
	var allRoutes6 []fortimgr.StaticRoute6
	var allDNS []fortimgr.DeviceDNS
	var allDDNS []fortimgr.DeviceDDNS
	var allIPAM []fortimgr.DeviceIPAM
	var allSDWAN []fortimgr.SDWANSettings
	var allSDWANMembers []fortimgr.SDWANMember
	var allSDWANServices []fortimgr.SDWANService
	var allSDWANDuplication []fortimgr.SDWANDuplication
	for _, d := range devices {
		dns, err := client.DeviceDNS(ctx, d.Name)
		if err != nil {
			fmt.Printf("  DNS(%s): %v\n", d.Name, err)
		} else if dns != nil {
			allDNS = append(allDNS, *dns)
		}

		ddns, err := client.ListDeviceDDNS(ctx, d.Name)
		if err != nil {
			fmt.Printf("  DDNS(%s): %v\n", d.Name, err)
		} else {
			allDDNS = append(allDDNS, ddns...)
		}

		ifaces, err := client.ListInterfaces(ctx, d.Name, "")
		if err != nil {
			fmt.Printf("  Interfaces(%s): %v\n", d.Name, err)
			continue
		}
		allInterfaces = append(allInterfaces, ifaces...)

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
			allVDOMNames = append(allVDOMNames, d.Name+"/"+v)

			routes, err := client.ListStaticRoutes(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  Routes(%s/%s): %v\n", d.Name, v, err)
				continue
			}
			allRoutes = append(allRoutes, routes...)

			routes6, err := client.ListStaticRoutes6(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  Routes6(%s/%s): %v\n", d.Name, v, err)
				continue
			}
			allRoutes6 = append(allRoutes6, routes6...)

			ipam, err := client.DeviceIPAM(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  IPAM(%s/%s): %v\n", d.Name, v, err)
			} else if ipam != nil {
				allIPAM = append(allIPAM, *ipam)
			}

			sdwan, err := client.SDWANSettings(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  SDWAN(%s/%s): %v\n", d.Name, v, err)
			} else if sdwan != nil {
				allSDWAN = append(allSDWAN, *sdwan)
			}

			sdwanMembers, err := client.ListSDWANMembers(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  SDWANMembers(%s/%s): %v\n", d.Name, v, err)
			} else {
				allSDWANMembers = append(allSDWANMembers, sdwanMembers...)
			}

			sdwanServices, err := client.ListSDWANServices(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  SDWANServices(%s/%s): %v\n", d.Name, v, err)
			} else {
				allSDWANServices = append(allSDWANServices, sdwanServices...)
			}

			sdwanDuplication, err := client.ListSDWANDuplication(ctx, d.Name, v)
			if err != nil {
				fmt.Printf("  SDWANDuplication(%s/%s): %v\n", d.Name, v, err)
			} else {
				allSDWANDuplication = append(allSDWANDuplication, sdwanDuplication...)
			}
		}
	}

	// System status.
	sysStatus, sysErr := client.SystemStatus(ctx)

	// Device firmware.
	firmware, fwErr := client.ListDeviceFirmware(ctx)
	firmwareUpgradePaths, firmwareUpgradePathErr := client.ListFirmwareUpgradePaths(ctx)
	psirtReport, psirtErr := client.DevicePSIRT(ctx, cfg.ADOM)

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
	fmt.Fprintf(w, "VDOMs\t%d\tOK\n", len(allVDOMNames))
	fmt.Fprintf(w, "Interfaces\t%d\tOK\n", len(allInterfaces))
	fmt.Fprintf(w, "Static Routes\t%d\tOK\n", len(allRoutes))
	fmt.Fprintf(w, "Static IPv6 Routes\t%d\tOK\n", len(allRoutes6))
	fmt.Fprintf(w, "Device DNS\t%d\tOK\n", len(allDNS))
	fmt.Fprintf(w, "Device DDNS\t%d\tOK\n", len(allDDNS))
	fmt.Fprintf(w, "Device IPAM\t%d\tOK\n", len(allIPAM))
	fmt.Fprintf(w, "SD-WAN Settings\t%d\tOK\n", len(allSDWAN))
	fmt.Fprintf(w, "SD-WAN Members\t%d\tOK\n", len(allSDWANMembers))
	fmt.Fprintf(w, "SD-WAN Services\t%d\tOK\n", len(allSDWANServices))
	fmt.Fprintf(w, "SD-WAN Duplication\t%d\tOK\n", len(allSDWANDuplication))
	if fwErr != nil {
		fmt.Fprintf(w, "Device Firmware\t0\t%s\n", fwErr)
	} else {
		fmt.Fprintf(w, "Device Firmware\t%d\tOK\n", len(firmware))
	}
	if firmwareUpgradePathErr != nil {
		fmt.Fprintf(w, "Firmware Upgrade Paths\t0\t%s\n", firmwareUpgradePathErr)
	} else {
		fmt.Fprintf(w, "Firmware Upgrade Paths\t%d\tOK\n", len(firmwareUpgradePaths))
	}
	if psirtErr != nil {
		fmt.Fprintf(w, "Device PSIRT\t0\t%s\n", psirtErr)
	} else {
		fmt.Fprintf(w, "Device PSIRT\t%d\tOK\n", len(psirtReport.ByIRNumber))
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
	writeSample("addresses6", firstN(addrs6, 10))
	writeSample("address_groups", firstN(groups, 5))
	writeSample("address_groups6", firstN(groups6, 5))
	writeSample("services", firstN(services, 10))
	writeSample("service_groups", firstN(svcGroups, 5))
	writeSample("schedules_recurring", firstN(recurring, 5))
	writeSample("schedules_onetime", firstN(onetime, 5))
	writeSample("schedule_groups", firstN(scheduleGroups, 5))
	writeSample("virtual_ips", firstN(virtualIPs, 5))
	writeSample("virtual_ip_groups", firstN(virtualIPGroups, 5))
	writeSample("virtual_ips6", firstN(virtualIPs6, 5))
	writeSample("virtual_ip_groups6", firstN(virtualIPGroups6, 5))
	writeSample("ip_pools", firstN(pools, 5))
	writeSample("ip_pools6", firstN(pools6, 5))
	writeSample("ip_pool_groups", firstN(poolGroups, 5))
	writeSample("internet_service_custom", firstN(internetServiceCustom, 5))
	writeSample("internet_service_custom_groups", firstN(internetServiceCustomGroups, 5))
	writeSample("internet_service_groups", firstN(internetServiceGroups, 5))
	writeSample("internet_service_names", firstN(internetServiceNames, 5))
	writeSample("fdsdb_internet_services", firstN(fdsdbInternetServices, 5))
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
	writeSample("adom_revisions", firstN(adomRevisions, 10))
	writeSample("workflow_sessions", firstN(workflowSessions, 10))
	writeSample("workflow_logs", firstN(workflowLogs, 10))
	writeSample("normalized_interfaces", firstN(normIfaces, 10))
	writeSample("device_assigned_packages", firstN(assignedPkgs, 10))
	if deviceSummary != nil {
		writeSample("device_summary", deviceSummary)
	}
	writeSample("vdom_names", allVDOMNames)
	writeSample("interfaces", firstN(allInterfaces, 10))
	writeSample("static_routes", firstN(allRoutes, 10))
	writeSample("static_routes6", firstN(allRoutes6, 10))
	writeSample("device_dns", firstN(allDNS, 10))
	writeSample("device_ddns", firstN(allDDNS, 10))
	writeSample("device_ipam", firstN(allIPAM, 10))
	writeSample("sdwan_settings", firstN(allSDWAN, 10))
	writeSample("sdwan_members", firstN(allSDWANMembers, 10))
	writeSample("sdwan_services", firstN(allSDWANServices, 10))
	writeSample("sdwan_duplication", firstN(allSDWANDuplication, 10))
	if sysStatus != nil {
		writeSample("system_status", sysStatus)
	}
	writeSample("device_firmware", firstN(firmware, 10))
	writeSample("firmware_upgrade_paths", firstN(firmwareUpgradePaths, 10))
	if psirtReport != nil {
		writeSample("device_psirt", psirtReport)
	}
}

// firstN returns the first n items from a slice (or all if fewer).
func firstN[T any](s []T, n int) []T {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
