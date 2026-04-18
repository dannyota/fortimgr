package fortimgr

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type apiDeviceFirmware struct {
	Name                  string `json:"name"`
	DevOID                int    `json:"devoid"`
	DevSN                 string `json:"devsn"`
	Model                 string `json:"model"`
	Platform              string `json:"platform_str"`
	PlatformID            int    `json:"platform_id"`
	OSType                int    `json:"os_type"`
	CurrVer               string `json:"curr_ver"`
	CurrBuild             int    `json:"curr_build"`
	UpdVer                string `json:"upd_ver"`
	UpdVerKey             string `json:"upd_ver_key"`
	KeyForDownloadRelease string `json:"key_for_download_release"`
	CanUpgrade            int    `json:"can_upgrade"`
	Connection            int    `json:"connection"`
	IsLicenseValid        int    `json:"is_license_valid"`
	IsModelDevice         int    `json:"is_model_device"`
	InvalidDate           string `json:"invalid_date"`
	UpgradeHistory        int    `json:"upgrade_history"`
	GroupName             string `json:"groupName"`
	Status                any    `json:"status"`
	IsGroup               int    `json:"isGroup"`
}

type apiDevicePSIRTReport struct {
	ByIRNumber        map[string]apiPSIRTAdvisory `json:"byIrNumber"`
	ByPlatform        map[string][]string         `json:"byPlatform"`
	NumDevicesPerRisk map[string]int              `json:"numDevicesPerRisk"`
}

type apiPSIRTAdvisory struct {
	IRNumber         string                               `json:"ir_number"`
	Title            string                               `json:"title"`
	Summary          string                               `json:"summary"`
	Description      string                               `json:"description"`
	Risk             int                                  `json:"risk"`
	ThreatSeverity   string                               `json:"threat_severity"`
	CVE              []string                             `json:"cve"`
	CVSS3            apiPSIRTCVSS3                        `json:"cvss3"`
	Products         map[string][]apiPSIRTProductVersion  `json:"products"`
	ImpactedProducts map[string][]apiPSIRTImpactedProduct `json:"impacted_products"`
}

type apiPSIRTCVSS3 struct {
	BaseScore     string `json:"cvss3_base_score"`
	ScoringVector string `json:"cvss3_scoring_vector"`
}

type apiPSIRTProductVersion struct {
	MinimumVersion string `json:"minimum_version"`
	MaximumVersion string `json:"maximum_version"`
	UpgradeTo      string `json:"upgrade_to"`
}

type apiPSIRTImpactedProduct struct {
	Major string `json:"major"`
	Minor string `json:"minor"`
	Patch string `json:"patch"`
}

// ListDeviceFirmware retrieves firmware status for all managed devices via flatui_proxy.
func (c *Client) ListDeviceFirmware(ctx context.Context) ([]DeviceFirmware, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	data, err := c.proxy(ctx, "/gui/adom/dvm/firmware/management", "loadFirmwareDataGroupByVersion")
	if err != nil {
		return nil, err
	}

	var items []apiDeviceFirmware
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal firmware: %w", err)
	}

	var result []DeviceFirmware
	for _, f := range items {
		if f.IsGroup == 1 {
			continue // skip group headers
		}
		result = append(result, DeviceFirmware{
			Name:               f.Name,
			DeviceOID:          f.DevOID,
			SerialNumber:       f.DevSN,
			Model:              f.Model,
			Platform:           f.Platform,
			PlatformID:         f.PlatformID,
			OSType:             f.OSType,
			CurrentVersion:     f.CurrVer,
			CurrentBuild:       f.CurrBuild,
			UpgradeVersion:     f.UpdVer,
			UpgradeVersionKey:  f.UpdVerKey,
			DownloadReleaseKey: f.KeyForDownloadRelease,
			CanUpgrade:         f.CanUpgrade == 1,
			Connected:          f.Connection == 1,
			LicenseValid:       f.IsLicenseValid == 1,
			ModelDevice:        f.IsModelDevice == 1,
			InvalidDate:        f.InvalidDate,
			UpgradeHistory:     f.UpgradeHistory,
			GroupName:          f.GroupName,
			Status:             toStringSlice(f.Status),
		})
	}

	return result, nil
}

// ListFirmwareUpgradePaths retrieves the firmware upgrade path matrix used by
// the Device Manager firmware view.
func (c *Client) ListFirmwareUpgradePaths(ctx context.Context) ([]FirmwareUpgradePath, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}

	data, err := c.proxy(ctx, "/gui/adom/dvm/device/firmware", "getUpdatePath")
	if err != nil {
		return nil, err
	}

	var rows []string
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal firmware upgrade paths: %w", err)
	}

	paths := make([]FirmwareUpgradePath, 0, len(rows))
	for _, row := range rows {
		paths = append(paths, parseFirmwareUpgradePath(row))
	}
	return paths, nil
}

func parseFirmwareUpgradePath(row string) FirmwareUpgradePath {
	fields := splitFirmwarePathRow(row)
	return FirmwareUpgradePath{
		Platform:        fields["FGTPlatform"],
		CurrentVersion:  fields["FGTCurrVersion"],
		CurrentBuild:    toInt(fields["FGTCurrBuildNum"]),
		UpgradeVersion:  fields["FGTUpgVersion"],
		UpgradeBuild:    toInt(fields["FGTUpgBuildNum"]),
		BaselineVersion: fields["BaselineVersion"],
		CurrentType:     fields["FGTCurrType"],
		UpgradeType:     fields["FGTUpgType"],
		CurrentEOES:     parseFirmwareDate(fields["FGTCurrEOES"]),
	}
}

func splitFirmwarePathRow(row string) map[string]string {
	result := map[string]string{}
	for _, part := range strings.Split(strings.TrimSpace(row), "|") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		result[key] = value
	}
	return result
}

func parseFirmwareDate(s string) time.Time {
	if len(s) != 8 {
		return time.Time{}
	}
	_, err := strconv.Atoi(s)
	if err != nil {
		return time.Time{}
	}
	t, err := time.Parse("20060102", s)
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}

// DevicePSIRT retrieves the Device Manager PSIRT/vulnerability summary for an
// ADOM.
func (c *Client) DevicePSIRT(ctx context.Context, adom string) (*DevicePSIRTReport, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	oid, err := c.adomOID(ctx, adom)
	if err != nil {
		return nil, err
	}
	data, err := c.proxy(ctx, fmt.Sprintf("/gui/adoms/%d/dvm/psirt", oid), "get")
	if err != nil {
		return nil, err
	}

	var raw apiDevicePSIRTReport
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("fortimgr: unmarshal PSIRT report: %w", err)
	}
	report := convertPSIRTReport(raw)
	return &report, nil
}

func convertPSIRTReport(raw apiDevicePSIRTReport) DevicePSIRTReport {
	advisories := make(map[string]PSIRTAdvisory, len(raw.ByIRNumber))
	for key, advisory := range raw.ByIRNumber {
		advisories[key] = convertPSIRTAdvisory(advisory)
	}
	return DevicePSIRTReport{
		ByIRNumber:        advisories,
		ByPlatform:        raw.ByPlatform,
		NumDevicesPerRisk: raw.NumDevicesPerRisk,
	}
}

func convertPSIRTAdvisory(raw apiPSIRTAdvisory) PSIRTAdvisory {
	return PSIRTAdvisory{
		IRNumber:         raw.IRNumber,
		Title:            raw.Title,
		Summary:          raw.Summary,
		Description:      raw.Description,
		Risk:             raw.Risk,
		ThreatSeverity:   raw.ThreatSeverity,
		CVE:              raw.CVE,
		CVSS3:            PSIRTCVSS3{BaseScore: raw.CVSS3.BaseScore, ScoringVector: raw.CVSS3.ScoringVector},
		Products:         convertPSIRTProducts(raw.Products),
		ImpactedProducts: convertPSIRTImpactedProducts(raw.ImpactedProducts),
	}
}

func convertPSIRTProducts(raw map[string][]apiPSIRTProductVersion) map[string][]PSIRTProductVersion {
	out := make(map[string][]PSIRTProductVersion, len(raw))
	for product, versions := range raw {
		out[product] = make([]PSIRTProductVersion, len(versions))
		for i, v := range versions {
			out[product][i] = PSIRTProductVersion(v)
		}
	}
	return out
}

func convertPSIRTImpactedProducts(raw map[string][]apiPSIRTImpactedProduct) map[string][]PSIRTImpactedProduct {
	out := make(map[string][]PSIRTImpactedProduct, len(raw))
	for product, versions := range raw {
		out[product] = make([]PSIRTImpactedProduct, len(versions))
		for i, v := range versions {
			out[product][i] = PSIRTImpactedProduct(v)
		}
	}
	return out
}
