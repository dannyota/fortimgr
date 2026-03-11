package fortimgr

// ADOM represents an Administrative Domain in FortiManager.
// ADOMs partition managed devices, policies, and objects into isolated scopes
// for multi-tenant or delegated management. Most single-tenant deployments
// use "root" (the default ADOM).
type ADOM struct {
	Name  string
	Desc  string
	State string // "enabled", "disabled"
	Mode  string // "normal", "backup"
	OSVer string // "7.0", "7.2", etc.
}

// Device represents a FortiGate device managed by FortiManager.
type Device struct {
	Name         string
	DeviceID     string
	SerialNumber string
	Platform     string
	Firmware     string // format: "7.2.5-b1517"
	HAMode       string // "standalone", "master", "slave"
	HAClusterID  string
	Status       string // "online", "offline"
	IPAddress    string
	ADOM         string
}

// PolicyPackage represents a policy package in FortiManager.
type PolicyPackage struct {
	Name  string
	Type  string
	ADOM  string
	Scope []string // device/group assignments
}

// Policy represents a firewall policy.
type Policy struct {
	PolicyID   int
	Name       string
	SrcIntf    []string
	DstIntf    []string
	SrcAddr    []string
	DstAddr    []string
	Service    []string
	Action     string
	Schedule   string
	NAT        string
	Status     string
	LogTraffic string
	Comments   string
}

// Address represents a firewall address object.
type Address struct {
	Name      string
	Type      string // ipmask, iprange, fqdn, geography, wildcard
	Subnet    string // CIDR notation
	StartIP   string
	EndIP     string
	FQDN      string
	Country   string
	Wildcard  string
	Comment   string
	Color     int
	AssocIntf string
}

// AddressGroup represents a firewall address group.
type AddressGroup struct {
	Name    string
	Members []string
	Comment string
	Color   int
}

// Service represents a firewall service object.
type Service struct {
	Name     string
	Protocol string
	TCPRange string
	UDPRange string
	Comment  string
}

// ServiceGroup represents a firewall service group.
type ServiceGroup struct {
	Name    string
	Members []string
	Comment string
}

// Schedule represents a firewall schedule.
type Schedule struct {
	Name  string
	Type  string // "recurring", "onetime"
	Day   string // recurring only, e.g. "monday tuesday"
	Start string
	End   string
	Color int
}

// VirtualIP represents a Virtual IP (VIP) object.
type VirtualIP struct {
	Name        string
	ExtIP       string
	MappedIP    string
	ExtIntf     string
	PortForward string
	Protocol    string
	ExtPort     string
	MappedPort  string
	Comment     string
	Color       int
}

// IPPool represents an IP Pool.
type IPPool struct {
	Name    string
	Type    string
	StartIP string
	EndIP   string
	Comment string
}

// Zone represents a system zone (interface grouping).
type Zone struct {
	Name        string
	Interfaces  []string
	Intrazone   string // "allow", "deny" — traffic policy between zone interfaces
	Description string
}

// VDOM represents a Virtual Domain on a FortiGate device.
type VDOM struct {
	Name   string
	Status string // "enable", "disable"
	OpMode string // "nat", "transparent"
}

// Interface represents a network interface on a FortiGate device.
type Interface struct {
	Name        string
	IP          string // CIDR notation
	Type        string // "physical", "vlan", "aggregate", "redundant", "tunnel", "wireless", "vdom-link", "loopback"
	Status      string // "up", "down"
	Role        string // "lan", "wan", "dmz", "undefined"
	Mode        string // "static", "dhcp", "pppoe"
	AllowAccess string // space-separated: "ping https ssh snmp http telnet fgfm"
	VDOM        string
	Zone        string
	VlanID      int
	MTU         int
	Speed       string
	Alias       string
	Description string
}

// StaticRoute represents a static route entry on a FortiGate device.
type StaticRoute struct {
	SeqNum   int
	Dst      string // CIDR notation
	Gateway  string
	Device   string // outgoing interface name
	Distance int
	Priority int
	Status   string // "enable", "disable"
	Comment  string
}

// AntivirusProfile represents an antivirus profile.
type AntivirusProfile struct {
	Name           string
	Comment        string
	ScanMode       string // "default", "legacy", "full"
	FeatureSet     string // "flow", "proxy"
	AVBlockLog     string // "enable", "disable"
	AVVirusLog     string // "enable", "disable"
	ExtendedLog    string // "enable", "disable"
	AnalyticsDB    string // "enable", "disable"
	MobileMalware  string // "enable", "disable"
}

// IPSSensor represents an IPS sensor.
type IPSSensor struct {
	Name                 string
	Comment              string
	ExtendedLog          string // "enable", "disable"
	BlockMaliciousURL    string // "enable", "disable"
	ScanBotnetConnections string // "disable", "block", "monitor"
}

// WebFilterProfile represents a web filter profile.
type WebFilterProfile struct {
	Name           string
	Comment        string
	FeatureSet     string // "flow", "proxy"
	InspectionMode string // "proxy", "flow-based", "dns"
	LogAllURL      string // "enable", "disable"
	WebContentLog  string // "enable", "disable"
	WebFTGDErrLog  string // "enable", "disable"
	ExtendedLog    string // "enable", "disable"
}

// AppControlProfile represents an application control profile.
type AppControlProfile struct {
	Name                   string
	Comment                string
	ExtendedLog            string // "enable", "disable"
	DeepAppInspection      string // "enable", "disable"
	EnforceDefaultAppPort  string // "enable", "disable"
	UnknownApplicationAction string // "pass", "block"
	UnknownApplicationLog    string // "enable", "disable"
	OtherApplicationAction   string // "pass", "block"
	OtherApplicationLog      string // "enable", "disable"
}

// SSLSSHProfile represents an SSL/SSH inspection profile.
type SSLSSHProfile struct {
	Name            string
	Comment         string
	ServerCertMode  string // "re-sign", "replace"
	MAPIOverHTTPS   string // "enable", "disable"
	RPCOverHTTPS    string // "enable", "disable"
	SSLAnomalyLog   string // "enable", "disable"
	SSLExemptionLog string // "enable", "disable"
	SupportedALPN   string // "none", "http1-1", "http2", "all"
}

// User represents a local user in FortiManager.
type User struct {
	Name   string
	Status string // "enable", "disable"
	Type   string // "local", "radius", "ldap", "tacacs+"
	Email  string
}

// UserGroup represents a user group.
type UserGroup struct {
	Name    string
	Members []string
	Type    string // "firewall", "fsso-service", "rsso", "guest"
	Comment string
}

// LDAPServer represents an LDAP server configuration.
// Credentials are intentionally excluded for security.
type LDAPServer struct {
	Name   string
	Server string
	Port   int
	DN     string
	Type   string // "simple", "anonymous", "regular"
	Secure string // "disable", "starttls", "ldaps"
}

// RADIUSServer represents a RADIUS server configuration.
// Credentials are intentionally excluded for security.
type RADIUSServer struct {
	Name     string
	Server   string
	AuthType string // "auto", "ms_chap_v2", "ms_chap", "chap", "pap"
	NASIP    string
}

// IPSecPhase1 represents an IPSec Phase 1 interface configuration.
type IPSecPhase1 struct {
	Name      string
	Interface string
	RemoteGW  string
	Proposal  string
	DHGroup   string
	Mode      string // "main", "aggressive"
	Type      string // "static", "dynamic", "ddns"
	Comments  string
}

// IPSecPhase2 represents an IPSec Phase 2 interface configuration.
type IPSecPhase2 struct {
	Name       string
	Phase1Name string
	Proposal   string
	SrcSubnet  string
	DstSubnet  string
	Comments   string
}

// SystemStatus represents FortiManager system status.
type SystemStatus struct {
	Hostname     string
	SerialNumber string
	Version      string // e.g. "v7.2.0-build1000 230101 (GA)"
	Platform     string
	Build        int
	HAMode       string // "standalone", "master", "slave"
}

// DeviceFirmware represents the firmware status of a managed device.
type DeviceFirmware struct {
	Name           string
	SerialNumber   string
	Platform       string
	CurrentVersion string
	CurrentBuild   int
	UpgradeVersion string
	CanUpgrade     bool
	Connected      bool
	LicenseValid   bool
}
