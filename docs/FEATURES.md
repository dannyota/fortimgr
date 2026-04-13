# Features

FlatUI resource coverage compared to FortiManager's official JSON-RPC API.

FlatUI uses the same internal API paths as JSON-RPC — the difference is the transport layer (session + CSRF + forward envelope vs token + direct). Some resources use the `flatui_proxy` endpoint instead (see [ARCHITECTURE.md](ARCHITECTURE.md)).

## 📄 Pagination (since v1.1.0)

Every `List*` method transparently fetches every page from FortiManager (1000 rows per forward request by default) and returns the concatenated result. Consumers never need to reason about offsets; deployments with thousands of policies, addresses, or revisions no longer risk silent truncation.

Two functional options are available on every paginated `List*` method:

```go
// Override the default 1000 rows-per-request page size.
addrs, _ := client.ListAddresses(ctx, "root", fortimgr.WithPageSize(500))

// Observe progress across pages.
revs, _ := client.ListADOMRevisions(ctx, "root",
    fortimgr.WithPageCallback(func(fetched, page int) {
        log.Printf("fetched %d revisions (page %d)", fetched, page)
    }),
)
```

**Three methods do NOT accept pagination options** — each for a different reason documented below. Every other `List*` method accepts `WithPageSize` / `WithPageCallback`.

| Method | Why no pagination |
|---|---|
| **`ListADOMs(ctx, all ...bool)`** | Existing `all ...bool` variadic parameter collides with `opts ...ListOption` (Go forbids two variadic parameters). ADOM count is hard-capped by the FortiManager license (~20–100 max on any deployment); paging is unnecessary. |
| **`ListDeviceFirmware(ctx)`** | Uses `flatui_proxy` with `loadFirmwareDataGroupByVersion`, which returns an aggregated grouping structure rather than a paginatable list. Parity against `ListDevices` was verified on small fleets; behavior on fleets of 500+ devices is unverified. |
| **`ListPackageInstallStatus(ctx, adom, pkg)`** | **FortiManager design limitation**: `/pm/config/adom/{adom}/_package/status` ignores the `range` parameter — passing `range:[999999,1]` still returns the full dataset. The endpoint always returns every row in one response, so accepting page options would be misleading. Callers that want a filtered result can still pass `pkg` to apply a server-side `filter` clause. |

## 🔐 Authentication

| Feature | JSON-RPC Endpoint | FlatUI Endpoint | Status |
|---------|-------------------|-----------------|:------:|
| Login | `EXEC /sys/login/user` | `POST /cgi-bin/module/flatui_auth` | Done |
| Logout | `EXEC /sys/logout` | `POST /cgi-bin/module/flatui_auth` | Done |
| Session auto-relogin | N/A (token-based) | Retry on code -6 | Done |
| API token auth | Supported | N/A (session only) | — |
| Multi-factor auth | Supported | Unknown | — |

## ⚙️ System & Administration

| Resource | SDK Method | API Endpoint | Transport | Status |
|----------|-----------|--------------|-----------|:------:|
| ADOMs | `ListADOMs()` | `/dvmdb/adom` | forward | Done |
| ADOM revision history | `ListADOMRevisions(adom)` | `/dvmdb/adom/{adom}/revision` | forward | Done |
| Workflow sessions | `ListWorkflowSessions(adom)` | `/dvmdb/adom/{adom}/workflow` | forward | Done |
| System status | `SystemStatus()` | `/gui/sys/config` | proxy | Done |
| Device firmware | `ListDeviceFirmware()` | `/gui/adom/dvm/firmware/management` | proxy | Done |
| HA cluster status | — | `/sys/ha/status` | — | — |
| Admin sessions | — | `/sys/session` | — | — |

## 🖥️ Device Management

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Devices | `ListDevices(adom)` | `/dvmdb/adom/{adom}/device` | Done¹ |
| VDOMs | `ListVDOMs(device)` | `/dvmdb/device/{device}/vdom` | Done |
| Interfaces | `ListInterfaces(device, vdom)` | `/pm/config/device/{device}/vdom/{vdom}/system/interface` | Done |
| Normalized interfaces | `ListNormalizedInterfaces(adom)` | `/pm/config/adom/{adom}/obj/dynamic/interface` | Done |
| Static routes | `ListStaticRoutes(device, vdom)` | `/pm/config/device/{device}/vdom/{vdom}/router/static` | Done |
| Zones | `ListZones(adom)` | `/pm/config/adom/{adom}/obj/system/zone` | Done |
| Device detail | — | `/dvmdb/device/{device}` | — |

Write operations (`add/device`, `del/device`) — not supported (read-only SDK).

¹ Since v1.0.3, `Device` carries extra sync-state fields from the same endpoint: `Hostname`, `ConfStatus` (`"unknown"` / `"insync"` / `"modified"`), `DevStatus` (`"auto_updated"` / `"installed"` / `"aborted"` / …), `LastChecked`, `LastResync`, `HARole`, and `HAMembers`. `HAMembers` is a `[]HAMember` slice that exposes every FortiGate inside an HA cluster (including the standby) — FortiManager models each HA cluster as one top-level device row with the primary's hostname, so `ListDevices` never returns passive members as separate rows; they only appear inside `HAMembers`. v1.1.0 adds 10 flat `License*` fields (`LicenseExpire`, `LicenseMaxCPU`, `LicenseRegion`, …) and switches the request to a server-side `fields` allowlist so encrypted device credentials (`adm_pass`, `private_key`, `psk`) never transit the wire. No activation key is exposed. See the godoc for the full list.

## 🛡️ Firewall Policy

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Policy packages | `ListPolicyPackages(adom)` | `/pm/pkg/adom/{adom}` | Done |
| Policies | `ListPolicies(adom, pkg)` | `/pm/config/adom/{adom}/pkg/{pkg}/firewall/policy` | Done |
| Package scope | Included in package response | Same | Done |
| Package install status | `ListPackageInstallStatus(adom, pkg)` | `/pm/config/adom/{adom}/_package/status` | Done |
| Policy hit count | — | `EXEC /sys/hitcount` | — |

Write operations (`SET/ADD/DELETE`) — not supported (read-only SDK).

## 🌐 Firewall Objects

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Addresses | `ListAddresses(adom)` | `/pm/config/adom/{adom}/obj/firewall/address` | Done |
| Address groups | `ListAddressGroups(adom)` | `/pm/config/adom/{adom}/obj/firewall/addrgrp` | Done |
| Services | `ListServices(adom)` | `/pm/config/adom/{adom}/obj/firewall/service/custom` | Done |
| Service groups | `ListServiceGroups(adom)` | `/pm/config/adom/{adom}/obj/firewall/service/group` | Done |
| Virtual IPs | `ListVirtualIPs(adom)` | `/pm/config/adom/{adom}/obj/firewall/vip` | Done |
| IP pools | `ListIPPools(adom)` | `/pm/config/adom/{adom}/obj/firewall/ippool` | Done |

### Address Types

| Type | Format |
|------|--------|
| `ipmask` | IP/CIDR (e.g. `10.0.0.0/24`) |
| `iprange` | Start-End IP |
| `fqdn` | Domain name |
| `geography` | Country code |
| `wildcard` | Wildcard mask |

## 📅 Scheduling

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Recurring schedules | `ListSchedulesRecurring(adom)` | `/pm/config/adom/{adom}/obj/firewall/schedule/recurring` | Done |
| One-time schedules | `ListSchedulesOnetime(adom)` | `/pm/config/adom/{adom}/obj/firewall/schedule/onetime` | Done |

## 🔒 Security Profiles

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Antivirus profiles | `ListAntivirusProfiles(adom)` | `/pm/config/adom/{adom}/obj/antivirus/profile` | Done |
| IPS sensors | `ListIPSSensors(adom)` | `/pm/config/adom/{adom}/obj/ips/sensor` | Done |
| Web filter profiles | `ListWebFilterProfiles(adom)` | `/pm/config/adom/{adom}/obj/webfilter/profile` | Done |
| App control profiles | `ListAppControlProfiles(adom)` | `/pm/config/adom/{adom}/obj/application/list` | Done |
| SSL/SSH profiles | `ListSSLSSHProfiles(adom)` | `/pm/config/adom/{adom}/obj/firewall/ssl-ssh-profile` | Done |

## 👤 User & Authentication

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Users | `ListUsers(adom)` | `/pm/config/adom/{adom}/obj/user/local` | Done |
| User groups | `ListUserGroups(adom)` | `/pm/config/adom/{adom}/obj/user/group` | Done |
| LDAP servers | `ListLDAPServers(adom)` | `/pm/config/adom/{adom}/obj/user/ldap` | Done |
| RADIUS servers | `ListRADIUSServers(adom)` | `/pm/config/adom/{adom}/obj/user/radius` | Done |

## 🔐 VPN

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| IPSec tunnels (phase 1) | `ListIPSecTunnels(adom)` / `ListIPSecPhase1(adom)` | `/pm/config/adom/{adom}/obj/vpn/ipsec/phase1-interface` | Done |
| IPSec selectors (phase 2) | `ListIPSecSelectors(adom)` / `ListIPSecPhase2(adom)` | `/pm/config/adom/{adom}/obj/vpn/ipsec/phase2-interface` | Done |
| SSL VPN settings | — | `/pm/config/.../vpn/ssl` | — |

## 📊 Logging & Monitoring — Future

| Resource | JSON-RPC Endpoint | Status |
|----------|-------------------|:------:|
| Log fetch | `/logview/adom/{adom}/logfiles/data` | — |
| Event alerts | Various | — |

## 📈 Summary

| Category | Done | Future | Total |
|----------|:----:|:------:|:-----:|
| Authentication | 3 | 2 | 5 |
| System & Administration | 3 | 2 | 5 |
| Device Management | 5 | 1 | 6 |
| Firewall Policy | 3 | 1 | 4 |
| Firewall Objects | 6 | 0 | 6 |
| Scheduling | 2 | 0 | 2 |
| Security Profiles | 5 | 0 | 5 |
| User & Authentication | 4 | 0 | 4 |
| VPN | 2 | 1 | 3 |
| Logging | 0 | 2 | 2 |
| **Total** | **33** | **9** | **42** |

## 📋 References

- [FortiManager JSON-RPC API](https://fndn.fortinet.net/) (official docs)
- [FortiManager Administration Guide](https://docs.fortinet.com/product/fortimanager/)
