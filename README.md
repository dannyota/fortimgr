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

Requires Go 1.23+.

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
defer client.Logout(ctx)

devices, err := client.ListDevices(ctx, "root")
if err != nil {
    log.Fatal(err)
}
for _, d := range devices {
    fmt.Printf("%s (%s) - %s\n", d.Name, d.SerialNumber, d.Status)
}
```

## 🛡️ Supported Resources

See [FEATURES.md](docs/FEATURES.md) for full coverage and JSON-RPC comparison.

| Category | Resources | Count |
|----------|-----------|:-----:|
| Device Management | Devices | 1 |
| Firewall Objects | Addresses, Address Groups, Services, Service Groups, VIPs, IP Pools | 6 |
| Policy | Policy Packages, Policies | 2 |
| Scheduling | Recurring Schedules, One-time Schedules | 2 |

## ⚠️ Known Issues

| Issue | Workaround |
|-------|------------|
| Non-RFC 5280 certs (negative serial numbers) | Set `GODEBUG=x509negativeserial=1` |
| No API versioning | Behavior may differ across FortiOS versions |

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE](docs/ARCHITECTURE.md) | FlatUI protocol, package layout, design decisions |
| [FEATURES](docs/FEATURES.md) | Resource coverage and JSON-RPC comparison |
| [PLAN](docs/PLAN.md) | Development roadmap |

## 📋 License

MIT — see [LICENSE](LICENSE).
