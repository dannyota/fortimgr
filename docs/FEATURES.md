# Features

FlatUI resource coverage compared to FortiManager's official JSON-RPC API.

FlatUI uses the same internal API paths as JSON-RPC — the difference is the transport layer (session + CSRF + forward envelope vs token + direct).

## 🔐 Authentication

| Feature | JSON-RPC Endpoint | FlatUI Endpoint | Status |
|---------|-------------------|-----------------|:------:|
| Login | `EXEC /sys/login/user` | `POST /cgi-bin/module/flatui_auth` | ✅ |
| Logout | `EXEC /sys/logout` | `POST /cgi-bin/module/flatui_auth` | ✅ |
| API token auth | Supported | N/A (session only) | — |
| Multi-factor auth | Supported | Unknown | — |

## 🖥️ Device Management

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Devices | `ListDevices()` | `/dvmdb/adom/{adom}/device` | ✅ |
| Device detail | — | `/dvmdb/device/{device}` | — |
| Firmware | Included in device response | Same | ✅ |
| HA status | Included in device response | Same | ✅ |

Write operations (`add/device`, `del/device`) — ❌ read-only SDK.

## 🛡️ Firewall Policy

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Policy packages | `ListPolicyPackages()` | `/pm/pkg/adom/{adom}` | ✅ |
| Policies | `ListPolicies()` | `/pm/config/adom/{adom}/pkg/{pkg}/firewall/policy` | ✅ |
| Package scope | Included in package response | Same | ✅ |
| Policy hit count | — | `EXEC /sys/hitcount` | — |

Write operations (`SET/ADD/DELETE`) — ❌ read-only SDK.

## 🌐 Firewall Objects

| Resource | SDK Method | API Endpoint | Status |
|----------|-----------|--------------|:------:|
| Addresses | `ListAddresses()` | `/pm/config/adom/{adom}/obj/firewall/address` | ✅ |
| Address groups | `ListAddressGroups()` | `/pm/config/adom/{adom}/obj/firewall/addrgrp` | ✅ |
| Services | `ListServices()` | `/pm/config/adom/{adom}/obj/firewall/service/custom` | ✅ |
| Service groups | `ListServiceGroups()` | `/pm/config/adom/{adom}/obj/firewall/service/group` | ✅ |
| Virtual IPs | `ListVirtualIPs()` | `/pm/config/adom/{adom}/obj/firewall/vip` | ✅ |
| IP pools | `ListIPPools()` | `/pm/config/adom/{adom}/obj/firewall/ippool` | ✅ |

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
| Recurring schedules | `ListSchedulesRecurring()` | `/pm/config/adom/{adom}/obj/firewall/schedule/recurring` | ✅ |
| One-time schedules | `ListSchedulesOnetime()` | `/pm/config/adom/{adom}/obj/firewall/schedule/onetime` | ✅ |

## ⚙️ System & Administration — Future

| Resource | JSON-RPC Endpoint | Status |
|----------|-------------------|:------:|
| ADOMs | `/dvmdb/adom` | — |
| System status | `/sys/status` | — |
| HA cluster status | `/sys/ha/status` | — |
| Admin sessions | `/sys/session` | — |

## 🔒 VPN — Future

| Resource | JSON-RPC Endpoint | Status |
|----------|-------------------|:------:|
| IPsec tunnels | `/pm/config/.../vpn.ipsec` | — |
| SSL VPN settings | `/pm/config/.../vpn.ssl` | — |

## 📊 Logging & Monitoring — Future

| Resource | JSON-RPC Endpoint | Status |
|----------|-------------------|:------:|
| Log fetch | `/logview/adom/{adom}/logfiles/data` | — |
| Event alerts | Various | — |

## 📈 Summary

| Category | ✅ Done | Future | Total |
|----------|:------:|:------:|:-----:|
| Authentication | 2 | 2 | 4 |
| Device Management | 3 | 1 | 4 |
| Firewall Policy | 3 | 1 | 4 |
| Firewall Objects | 6 | 0 | 6 |
| Scheduling | 2 | 0 | 2 |
| System | 0 | 4 | 4 |
| VPN | 0 | 2 | 2 |
| Logging | 0 | 2 | 2 |
| **Total** | **16** | **12** | **28** |

## 📋 References

- [FortiManager JSON-RPC API](https://fndn.fortinet.net/) (official docs)
- [FortiManager Administration Guide](https://docs.fortinet.com/product/fortimanager/)
