package fortimgr

import (
	"encoding/json"
	"time"
)

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
//
// Note on HA fields: HAMode is the legacy field and still uses the historical
// "standalone" / "master" / "slave" mapping for backwards compatibility.
// New code should read HATopology for cluster topology and HARole for the
// device's role.
type Device struct {
	Name         string
	DeviceID     string
	SerialNumber string
	Platform     string
	Firmware     string // format: "7.2.5-b1517"
	HAMode       string // "standalone", "master", "slave" — legacy, kept for compat
	HATopology   string // "standalone", "active-passive", "active-active"
	HAClusterID  string
	Status       string // "online", "offline"
	IPAddress    string
	ADOM         string

	// v1.0.3 additions — populated from /dvmdb/adom/{adom}/device.
	Hostname    string    // hostname as reported by the FortiGate
	ConfStatus  string    // "unknown", "insync", "modified"
	DevStatus   string    // "none", "unknown", "auto_updated", "aborted", ... (see devStatuses)
	LastChecked time.Time // when FortiManager last heard from the device; zero if never
	LastResync  time.Time // when FortiManager last resync'd device config; zero if never
	HARole      string    // "", "master", "slave" — role of the top-level record

	// HAMembers holds every HA cluster member the FortiManager knows about for
	// this device record. FortiManager models each HA cluster as a single
	// top-level Device entry (with Name/Hostname set to the primary), so the
	// secondary FortiGates in an active-passive pair never appear as standalone
	// Device rows. HAMembers is the only place where standby members surface.
	//
	// For standalone FortiGates HAMembers is empty.
	HAMembers []HAMember

	// v1.1.0 license / subscription metadata.
	//
	// Populated from the same /dvmdb/adom/{adom}/device endpoint via a
	// server-side fields allowlist, so no encrypted credentials transit
	// the wire. The VMLicense* fields apply only to VM-licensed
	// FortiGates — they stay at their zero values (0, false, or
	// time.Time{}) for hardware appliances. The other License* fields
	// are populated on all devices.
	//
	// None of these fields carries the license activation key.
	// FortiManager does not persist activation keys; only status,
	// capacity, expiry, and region metadata.
	LicenseExpire       time.Time // VM license expiry; zero = perpetual / not licensed
	LicenseOverdueSince time.Time // when VM license went overdue; zero = never
	LicenseMaxCPU       int       // max vCPU permitted by the VM license
	LicenseMaxRAM       int       // max RAM (MB) permitted by the VM license
	LicenseUTMEnabled   bool      // whether the VM license includes the UTM bundle
	LicenseType         int       // raw VM license type enum, passed through
	LicenseInstalledAt  time.Time // when the VM license was installed; zero if never
	LicenseLastSync     time.Time // last FortiGuard sync; zero if never
	LicenseRegion       string    // license region (lic_region)
	LicenseFlags        int       // license bitmask (lic_flags), passed through
}

// HAMember describes one FortiGate inside an HA cluster, as returned in the
// ha_slave array of /dvmdb/adom/{adom}/device. This is how the SDK exposes
// passive/secondary members — they do not appear as top-level Device entries
// in ListDevices.
type HAMember struct {
	Name         string
	SerialNumber string
	Role         string // "master", "slave"
	Status       string // "online", "offline"
	ConfStatus   string // "unknown", "insync", "modified"
}

// AssignedPackageRef describes a policy/profile package reference assigned to
// a managed device or VDOM.
type AssignedPackageRef struct {
	Name   string
	OID    int
	Flags  int
	Status int
}

// DeviceAssignedPackage describes the package/profile assignment FortiManager
// shows in Device Manager for one device or VDOM.
type DeviceAssignedPackage struct {
	DeviceOID    int
	VDOMOID      int
	Package      AssignedPackageRef
	FAPProfile   AssignedPackageRef
	FExtProfile  AssignedPackageRef
	ProfileDirty bool
}

// DeviceSummary contains the install/configuration context FortiManager shows
// in the device dashboard. Fields stay at zero values when the current
// FortiManager version omits them.
type DeviceSummary struct {
	ADOM                  string
	Device                string
	Hostname              string
	SerialNumber          string
	Firmware              string
	ConfigStatus          string
	TotalRevisions        int
	LastInstalledRevision int
	LastInstallation      string
	LastInstallTime       time.Time
	InstalledBy           string
	HAMode                string
	HAUpgradeMode         string
	HAClusterName         string
	HAClusterID           int
	HAMembers             []DeviceSummaryHAMember
}

// DeviceSummaryHAMember describes one HA member as shown by the device
// dashboard summary.
type DeviceSummaryHAMember struct {
	OID          int
	Name         string
	SerialNumber string
	Role         string
	Status       int
	SyncStatus   string
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

// Address6 represents an IPv6 firewall address object.
type Address6 struct {
	Name      string
	Type      string
	IP6       string
	Comment   string
	Color     int
	AssocIntf string
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

type VirtualIP6 struct {
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

type VirtualIPGroup struct {
	Name    string
	Members []string
	Comment string
	Color   int
}

// IPPool represents an IP Pool.
type IPPool struct {
	Name    string
	Type    string
	StartIP string
	EndIP   string
	Comment string
}

type IPPool6 struct {
	Name    string
	Type    string
	StartIP string
	EndIP   string
	Comment string
}

type IPPoolGroup struct {
	Name    string
	Members []string
	Comment string
}

type InternetServiceCustom struct {
	Name    string
	ID      int
	Comment string
	Entry   []string
}

type InternetServiceGroup struct {
	Name    string
	ID      int
	Members []string
	Comment string
}

type InternetServiceName struct {
	Name string
	ID   int
}

type FDSDBInternetService struct {
	Name     string
	ID       int
	Category string
	Protocol string
}

type ScheduleGroup struct {
	Name    string
	Members []string
	Comment string
	Color   int
}

// Zone represents a system zone (interface grouping).
type Zone struct {
	Name        string
	Interfaces  []string
	Intrazone   string // "allow", "deny" — traffic policy between zone interfaces
	Description string
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

// StaticRoute6 represents an IPv6 static route entry on a FortiGate device.
type StaticRoute6 struct {
	SeqNum   int
	Dst      string
	Gateway  string
	Device   string
	Distance int
	Priority int
	Status   string // "enable", "disable"
	Comment  string
}

// DeviceDNS represents global DNS settings on a managed FortiGate.
type DeviceDNS struct {
	OID                    int
	Primary                string
	Secondary              string
	AltPrimary             string
	AltSecondary           string
	IPv6Primary            string
	IPv6Secondary          string
	Protocol               string
	ServerSelectMethod     string
	ServerHostname         []string
	Domain                 []string
	InterfaceSelectMethod  string
	Interfaces             []string
	SourceIP               string
	Retry                  int
	Timeout                int
	DNSCacheLimit          int
	DNSCacheTTL            int
	FQDNCacheTTL           int
	FQDNMinRefresh         int
	FQDNMaxRefresh         int
	CacheNotFoundResponses string
	Log                    string
	SSLCertificate         []string
}

// DeviceDDNS represents one DDNS entry from a managed FortiGate.
type DeviceDDNS struct {
	Name             string
	DDNSDomain       string
	MonitorInterface string
	Server           string
	Status           string
	Username         string
}

// DeviceIPAM represents IPAM settings on a managed FortiGate VDOM.
type DeviceIPAM struct {
	OID                         int
	Status                      string
	ServerType                  string
	AutomaticConflictResolution string
	ManageLANAddresses          string
	ManageLANExtensionAddresses string
	ManageSSIDAddresses         string
	RequireSubnetSizeMatch      string
}

// SDWANSettings represents SD-WAN settings on a managed FortiGate VDOM.
// Secret fields such as health-check passwords are intentionally not exposed.
type SDWANSettings struct {
	OID                    int
	Status                 string
	LoadBalanceMode        string
	FailDetect             string
	AppPerfLogPeriod       int
	DuplicationMaxNum      int
	NeighborHoldBootTime   int
	NeighborHoldDown       string
	NeighborHoldDownTime   int
	SpeedtestBypassRouting string
	FailAlertInterfaces    []string
	HealthChecks           []SDWANHealthCheck
	Zones                  []SDWANZone
}

type SDWANHealthCheck struct {
	Name                   string
	OID                    int
	Protocol               string
	Server                 []string
	Members                []int
	Interval               int
	Failtime               int
	Recoverytime           int
	Source                 string
	Source6                string
	SystemDNS              string
	UpdateStaticRoute      string
	UpdateCascadeInterface string
	SLA                    []SDWANSLA
}

type SDWANSLA struct {
	ID                  int
	OID                 int
	LatencyThreshold    int
	JitterThreshold     int
	PacketLossThreshold int
	LinkCostFactor      string
}

type SDWANZone struct {
	Name                  string
	OID                   int
	MinimumSLAMeetMembers int
	ServiceSLATieBreak    string
	ADVPNSelect           string
}

type SDWANMember struct {
	SeqNum    int
	Interface string
	Gateway   string
	Gateway6  string
	Zone      string
	Status    string
	Cost      int
	Priority  int
	Comment   string
}

type SDWANService struct {
	ID              int
	Name            string
	Mode            string
	Status          string
	AddressMode     string
	InputDevice     []string
	Source          []string
	Destination     []string
	InternetService string
	PriorityMembers []string
}

type SDWANDuplication struct {
	ID                int
	Name              string
	ServiceID         int
	PacketDuplication string
	Fields            []string
}

// AntivirusProfile represents an antivirus profile.
type AntivirusProfile struct {
	Name          string
	Comment       string
	ScanMode      string // "default", "legacy", "full"
	FeatureSet    string // "flow", "proxy"
	AVBlockLog    string // "enable", "disable"
	AVVirusLog    string // "enable", "disable"
	ExtendedLog   string // "enable", "disable"
	AnalyticsDB   string // "enable", "disable"
	MobileMalware string // "enable", "disable"
}

// IPSSensor represents an IPS sensor.
type IPSSensor struct {
	Name                  string
	Comment               string
	ExtendedLog           string // "enable", "disable"
	BlockMaliciousURL     string // "enable", "disable"
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
	Name                     string
	Comment                  string
	ExtendedLog              string // "enable", "disable"
	DeepAppInspection        string // "enable", "disable"
	EnforceDefaultAppPort    string // "enable", "disable"
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

// IPSecSelector represents an IPSec quick-mode selector, the friendlier
// FortiGate-GUI-aligned view of a Phase 2 entry. Each selector is bound
// to a parent tunnel via the Tunnel field.
type IPSecSelector struct {
	Name      string
	Tunnel    string // parent IPSec tunnel (phase 1) name
	Proposal  string
	SrcSubnet string
	DstSubnet string
	Comments  string
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
	Name               string
	DeviceOID          int
	SerialNumber       string
	Model              string
	Platform           string
	PlatformID         int
	OSType             int
	CurrentVersion     string
	CurrentBuild       int
	UpgradeVersion     string
	UpgradeVersionKey  string
	DownloadReleaseKey string
	CanUpgrade         bool
	Connected          bool
	LicenseValid       bool
	ModelDevice        bool
	InvalidDate        string
	UpgradeHistory     int
	GroupName          string
	Status             []string
}

// FirmwareUpgradePath represents one available firmware upgrade transition.
type FirmwareUpgradePath struct {
	Platform        string
	CurrentVersion  string
	CurrentBuild    int
	UpgradeVersion  string
	UpgradeBuild    int
	BaselineVersion string
	CurrentType     string
	UpgradeType     string
	CurrentEOES     time.Time
}

// DevicePSIRTReport is the firmware vulnerability summary returned by the
// Device Manager PSIRT view.
type DevicePSIRTReport struct {
	ByIRNumber        map[string]PSIRTAdvisory
	ByPlatform        map[string][]string
	NumDevicesPerRisk map[string]int
}

type PSIRTAdvisory struct {
	IRNumber         string
	Title            string
	Summary          string
	Description      string
	Risk             int
	ThreatSeverity   string
	CVE              []string
	CVSS3            PSIRTCVSS3
	Products         map[string][]PSIRTProductVersion
	ImpactedProducts map[string][]PSIRTImpactedProduct
}

type PSIRTCVSS3 struct {
	BaseScore     string
	ScoringVector string
}

type PSIRTProductVersion struct {
	MinimumVersion string
	MaximumVersion string
	UpgradeTo      string
}

type PSIRTImpactedProduct struct {
	Major string
	Minor string
	Patch string
}

// ADOMRevision represents one entry in an ADOM's revision history, as
// returned by /dvmdb/adom/{adom}/revision. Each revision is a snapshot
// created when a change was applied (by workflow submission, install
// preview, or manual revision). Revisions join against WorkflowSession
// via the revid field, and against PackageInstallStatus for the "what
// was installed when" silver view.
type ADOMRevision struct {
	Version   int
	Name      string
	Desc      string
	CreatedBy string
	CreatedAt time.Time
	Locked    bool
}

// WorkflowSession represents one change request in an ADOM's workflow
// history, as returned by /dvmdb/adom/{adom}/workflow. A workflow
// session tracks who created a change, who submitted it for approval,
// who audited/approved it, and which revision it produced. It is the
// primary audit trail for change management.
//
// The State field is best-effort — FortiManager documentation does not
// list the complete set of state values, so unknown ints are passed
// through unchanged. Sessions with create, submit, and audit timestamps
// all populated are "approved" (state 3 observed empirically).
type WorkflowSession struct {
	SessionID   int
	Name        string
	Description string
	State       string
	CreatedBy   string
	CreatedAt   time.Time
	SubmittedBy string
	SubmittedAt time.Time
	AuditedBy   string
	AuditedAt   time.Time
	RevisionID  int // joins to ADOMRevision.Version
}

// WorkflowLog represents one log line inside a workflow session, as returned by
// /dvmdb/adom/{adom}/workflow/{sessionID}/wflog.
type WorkflowLog struct {
	OID         int
	Sequence    int
	SessionID   int
	Action      int
	User        string
	Timestamp   time.Time
	Description string
	Flags       int
}

// PolicyRevision represents one entry in a firewall policy's per-object
// revision history, as returned by
// /pm/config/adom/{adom}/_objrev/pkg/{pkg}/firewall/policy/{id}.
//
// Each revision records a single change to the policy: who made it,
// when, what action was taken (Action + Note), and the full policy
// snapshot at that point in time (Config). Revisions are numbered from
// 1 (oldest / initial creation) and ordered oldest-first.
type PolicyRevision struct {
	Revision  int
	Action    string // "add", "modify", or raw int for unmapped values
	Note      string
	User      string
	Timestamp time.Time
	PolicyID  int
	OID       int
	Config    json.RawMessage
}

// NormalizedInterface represents a FortiManager normalized interface —
// an ADOM-level interface abstraction that policies reference by name.
// FortiManager maps each normalized name to one or more per-device
// physical interfaces via the Mappings slice, so a single policy using
// "wan1" can apply to different physical interfaces on different
// FortiGates in the same ADOM.
//
// Returned by /pm/config/adom/{adom}/obj/dynamic/interface.
type NormalizedInterface struct {
	Name           string
	SingleIntf     bool
	ZoneOnly       bool
	Wildcard       bool
	DefaultMapping bool
	Color          int

	// Mappings holds one entry per {device, vdom} scope the normalized
	// name resolves to. A normalized interface with N scopes in its
	// raw FMG dynamic_mapping array produces N Mappings entries — the
	// SDK flattens the nested _scope array into individual rows for
	// easier downstream iteration.
	//
	// For declared-but-unmapped normalized interfaces (the majority
	// on most ADOMs), Mappings is empty.
	Mappings []NormalizedInterfaceMapping
}

// NormalizedInterfaceMapping is one {device, vdom} scope of a
// NormalizedInterface, naming the concrete physical interfaces on
// that device/VDOM that the normalized name resolves to.
type NormalizedInterfaceMapping struct {
	Device        string
	VDOM          string
	LocalIntf     []string // local interface names on the mapped device
	IntrazoneDeny bool
}

// PackageInstallStatus represents the install state of a policy package on
// a single device/VDOM target. The Status field distinguishes assignment
// (device is on the scope list) from actual installation (config has been
// pushed and is running on the FortiGate) — callers that want to verify
// "this policy is actually enforcing" should check Status == "installed".
//
// The underlying /pm/config/adom/{adom}/_package/status endpoint does not
// expose revision numbers, install time, or modify state. Callers that need
// those should join against ADOM revision history (landing in v1.1.0).
type PackageInstallStatus struct {
	ADOM    string
	Package string
	Device  string
	VDOM    string
	Status  string // "installed", "modified", "never", "unknown", "imported"
}
