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
