# fortimgr

Unofficial Go SDK for FortiManager using the FlatUI (Web UI) API.

> **Use at your own risk.** FlatUI is an undocumented internal API. It may break
> with any firmware update. Fortinet does not support or endorse this library.

## ⚠️ When to Use

| Scenario | Recommendation |
|----------|----------------|
| JSON-RPC API available | Use the [official API](https://fndn.fortinet.net/) instead |
| JSON-RPC disabled (hardened env) | Use this SDK |
| Write operations needed | Not supported (read-only SDK) |

## 📚 Install

```bash
go get danny.vn/fortimgr
```

Requires Go 1.24+.

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "danny.vn/fortimgr"
)

func main() {
    client, err := fortimgr.NewClient("https://fortimanager.example.com",
        fortimgr.WithCredentials("admin", "password"),
        fortimgr.WithInsecureTLS(), // self-signed certs
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()
    if err := client.Login(ctx); err != nil {
        log.Fatal(err)
    }

    // "root" is the default ADOM (Administrative Domain).
    devices, err := client.ListDevices(ctx, "root")
    if err != nil {
        log.Fatal(err)
    }
    for _, d := range devices {
        fmt.Printf("%s (%s) - %s\n", d.Name, d.SerialNumber, d.Status)
    }
}
```

## 🏢 ADOMs

Every `List*` method takes an `adom` parameter — this is a FortiManager
**Administrative Domain** (ADOM). ADOMs partition managed devices, policies, and
objects into isolated scopes for multi-tenant or delegated management.

Most single-tenant deployments use `"root"` (the default ADOM).

## 🛡️ Supported Resources

See [FEATURES.md](docs/FEATURES.md) for full coverage and JSON-RPC comparison.

| Category | Resources | Count |
|----------|-----------|:-----:|
| Administration | ADOMs | 1 |
| Device Management | Devices | 1 |
| Firewall Objects | Addresses, Address Groups, Services, Service Groups, Virtual IPs, IP Pools | 6 |
| Policy | Policy Packages, Policies | 2 |
| Scheduling | Recurring Schedules, One-time Schedules | 2 |

## ✅ Testing

```bash
go test ./...
```

Smoke test against a live FortiManager:

```bash
FORTIMGR_ADDRESS=https://fm.example.com \
FORTIMGR_USERNAME=admin \
FORTIMGR_PASSWORD=secret \
go run smoke.go
```

## ⚠️ Known Issues

| Issue | Workaround |
|-------|------------|
| Non-RFC 5280 certs (negative serial numbers) | Use `WithX509NegativeSerial()` |
| No API versioning | Behavior may differ across FortiOS versions |

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE](docs/ARCHITECTURE.md) | FlatUI protocol, package layout, design decisions |
| [FEATURES](docs/FEATURES.md) | Resource coverage and JSON-RPC comparison |
| [PLAN](docs/PLAN.md) | Development roadmap |

## 📋 License

MIT — see [LICENSE](LICENSE).
