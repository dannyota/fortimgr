# Features

FlatUI resource coverage compared to FortiManager's official JSON-RPC API.

FlatUI uses the same internal API paths as JSON-RPC — the difference is the transport layer (session + CSRF + forward envelope vs token + direct). Some resources use the `flatui_proxy` endpoint instead (see [ARCHITECTURE.md](ARCHITECTURE.md)).

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
| System status | `SystemStatus()` | `/gui/sys/config` | proxy | Done |
| Device firmware | `ListDeviceFirmware()` | `/gui/adom/dvm/firmware/management` | proxy | Done |
| HA cluster status | — | `/sys/ha/status` | — | — |
| Admin sessions | — | `/sys/session` | — | — |

## 🖥️ Device Management

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Devices | `ListDevices(adom)` | `/dvmdb/adom/{adom}/device` | Done |
| VDOMs | `ListVDOMs(device)` | `/dvmdb/device/{device}/vdom` | Done |
| Interfaces | `ListInterfaces(device, vdom)` | `/pm/config/device/{device}/vdom/{vdom}/system/interface` | Done |
| Static routes | `ListStaticRoutes(device, vdom)` | `/pm/config/device/{device}/vdom/{vdom}/router/static` | Done |
| Zones | `ListZones(adom)` | `/pm/config/adom/{adom}/obj/system/zone` | Done |
| Device detail | — | `/dvmdb/device/{device}` | — |

Write operations (`add/device`, `del/device`) — not supported (read-only SDK).

## 🛡️ Firewall Policy

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Policy packages | `ListPolicyPackages(adom)` | `/pm/pkg/adom/{adom}` | Done |
| Policies | `ListPolicies(adom, pkg)` | `/pm/config/adom/{adom}/pkg/{pkg}/firewall/policy` | Done |
| Package scope | Included in package response | Same | Done |
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
| IPSec Phase 1 | `ListIPSecPhase1(adom)` | `/pm/config/adom/{adom}/obj/vpn.ipsec/phase1-interface` | Done |
| IPSec Phase 2 | `ListIPSecPhase2(adom)` | `/pm/config/adom/{adom}/obj/vpn.ipsec/phase2-interface` | Done |
| SSL VPN settings | — | `/pm/config/.../vpn.ssl` | — |

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
