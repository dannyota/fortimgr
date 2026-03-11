package fortimgr

import (
	"context"
	"fmt"
)

type apiAntivirusProfile struct {
	Name          string `json:"name"`
	Comment       string `json:"comment"`
	ScanMode      any    `json:"scan-mode"`
	FeatureSet    any    `json:"feature-set"`
	AVBlockLog    any    `json:"av-block-log"`
	AVVirusLog    any    `json:"av-virus-log"`
	ExtendedLog   any    `json:"extended-log"`
	AnalyticsDB   any    `json:"analytics-db"`
	MobileMalware any    `json:"mobile-malware-db"`
}

type apiIPSSensor struct {
	Name                  string `json:"name"`
	Comment               string `json:"comment"`
	ExtendedLog           any    `json:"extended-log"`
	BlockMaliciousURL     any    `json:"block-malicious-url"`
	ScanBotnetConnections any    `json:"scan-botnet-connections"`
}

type apiWebFilterProfile struct {
	Name           string `json:"name"`
	Comment        string `json:"comment"`
	FeatureSet     any    `json:"feature-set"`
	InspectionMode any    `json:"inspection-mode"`
	LogAllURL      any    `json:"log-all-url"`
	WebContentLog  any    `json:"web-content-log"`
	WebFTGDErrLog  any    `json:"web-ftgd-err-log"`
	ExtendedLog    any    `json:"extended-log"`
}

type apiAppControlProfile struct {
	Name                     string `json:"name"`
	Comment                  string `json:"comment"`
	ExtendedLog              any    `json:"extended-log"`
	DeepAppInspection        any    `json:"deep-app-inspection"`
	EnforceDefaultAppPort    any    `json:"enforce-default-app-port"`
	UnknownApplicationAction any    `json:"unknown-application-action"`
	UnknownApplicationLog    any    `json:"unknown-application-log"`
	OtherApplicationAction   any    `json:"other-application-action"`
	OtherApplicationLog      any    `json:"other-application-log"`
}

type apiSSLSSHProfile struct {
	Name            string `json:"name"`
	Comment         string `json:"comment"`
	ServerCertMode  any    `json:"server-cert-mode"`
	MAPIOverHTTPS   any    `json:"mapi-over-https"`
	RPCOverHTTPS    any    `json:"rpc-over-https"`
	SSLAnomalyLog   any    `json:"ssl-anomaly-log"`
	SSLExemptionLog any    `json:"ssl-exemption-log"`
	SupportedALPN   any    `json:"supported-alpn"`
}

// ListAntivirusProfiles retrieves antivirus profiles from an ADOM.
func (c *Client) ListAntivirusProfiles(ctx context.Context, adom string) ([]AntivirusProfile, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/antivirus/profile", adom)
	items, err := get[apiAntivirusProfile](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	profiles := make([]AntivirusProfile, len(items))
	for i, p := range items {
		profiles[i] = AntivirusProfile{
			Name:          p.Name,
			Comment:       p.Comment,
			ScanMode:      mapEnum(toString(p.ScanMode), scanModes),
			FeatureSet:    mapEnum(toString(p.FeatureSet), featureSets),
			AVBlockLog:    mapEnum(toString(p.AVBlockLog), enableDisable),
			AVVirusLog:    mapEnum(toString(p.AVVirusLog), enableDisable),
			ExtendedLog:   mapEnum(toString(p.ExtendedLog), enableDisable),
			AnalyticsDB:   mapEnum(toString(p.AnalyticsDB), enableDisable),
			MobileMalware: mapEnum(toString(p.MobileMalware), enableDisable),
		}
	}

	return profiles, nil
}

// ListIPSSensors retrieves IPS sensors from an ADOM.
func (c *Client) ListIPSSensors(ctx context.Context, adom string) ([]IPSSensor, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/ips/sensor", adom)
	items, err := get[apiIPSSensor](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	sensors := make([]IPSSensor, len(items))
	for i, s := range items {
		sensors[i] = IPSSensor{
			Name:                  s.Name,
			Comment:               s.Comment,
			ExtendedLog:           mapEnum(toString(s.ExtendedLog), enableDisable),
			BlockMaliciousURL:     mapEnum(toString(s.BlockMaliciousURL), enableDisable),
			ScanBotnetConnections: mapEnum(toString(s.ScanBotnetConnections), botnetConnections),
		}
	}

	return sensors, nil
}

// ListWebFilterProfiles retrieves web filter profiles from an ADOM.
func (c *Client) ListWebFilterProfiles(ctx context.Context, adom string) ([]WebFilterProfile, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/webfilter/profile", adom)
	items, err := get[apiWebFilterProfile](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	profiles := make([]WebFilterProfile, len(items))
	for i, p := range items {
		profiles[i] = WebFilterProfile{
			Name:           p.Name,
			Comment:        p.Comment,
			FeatureSet:     mapEnum(toString(p.FeatureSet), featureSets),
			InspectionMode: mapEnum(toString(p.InspectionMode), inspectionModes),
			LogAllURL:      mapEnum(toString(p.LogAllURL), enableDisable),
			WebContentLog:  mapEnum(toString(p.WebContentLog), enableDisable),
			WebFTGDErrLog:  mapEnum(toString(p.WebFTGDErrLog), enableDisable),
			ExtendedLog:    mapEnum(toString(p.ExtendedLog), enableDisable),
		}
	}

	return profiles, nil
}

// ListAppControlProfiles retrieves application control profiles from an ADOM.
func (c *Client) ListAppControlProfiles(ctx context.Context, adom string) ([]AppControlProfile, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/application/list", adom)
	items, err := get[apiAppControlProfile](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	profiles := make([]AppControlProfile, len(items))
	for i, p := range items {
		profiles[i] = AppControlProfile{
			Name:                     p.Name,
			Comment:                  p.Comment,
			ExtendedLog:              mapEnum(toString(p.ExtendedLog), enableDisable),
			DeepAppInspection:        mapEnum(toString(p.DeepAppInspection), enableDisable),
			EnforceDefaultAppPort:    mapEnum(toString(p.EnforceDefaultAppPort), enableDisable),
			UnknownApplicationAction: mapEnum(toString(p.UnknownApplicationAction), unknownAppActions),
			UnknownApplicationLog:    mapEnum(toString(p.UnknownApplicationLog), enableDisable),
			OtherApplicationAction:   mapEnum(toString(p.OtherApplicationAction), unknownAppActions),
			OtherApplicationLog:      mapEnum(toString(p.OtherApplicationLog), enableDisable),
		}
	}

	return profiles, nil
}

// ListSSLSSHProfiles retrieves SSL/SSH inspection profiles from an ADOM.
func (c *Client) ListSSLSSHProfiles(ctx context.Context, adom string) ([]SSLSSHProfile, error) {
	if !c.LoggedIn() {
		return nil, ErrNotLoggedIn
	}
	if !validName(adom) {
		return nil, fmt.Errorf("%w: %q", ErrInvalidName, adom)
	}

	apiURL := fmt.Sprintf("/pm/config/adom/%s/obj/firewall/ssl-ssh-profile", adom)
	items, err := get[apiSSLSSHProfile](ctx, c, apiURL)
	if err != nil {
		return nil, err
	}

	profiles := make([]SSLSSHProfile, len(items))
	for i, p := range items {
		profiles[i] = SSLSSHProfile{
			Name:            p.Name,
			Comment:         p.Comment,
			ServerCertMode:  mapEnum(toString(p.ServerCertMode), serverCertModes),
			MAPIOverHTTPS:   mapEnum(toString(p.MAPIOverHTTPS), enableDisable),
			RPCOverHTTPS:    mapEnum(toString(p.RPCOverHTTPS), enableDisable),
			SSLAnomalyLog:   mapEnum(toString(p.SSLAnomalyLog), enableDisable),
			SSLExemptionLog: mapEnum(toString(p.SSLExemptionLog), enableDisable),
			SupportedALPN:   mapEnum(toString(p.SupportedALPN), supportedALPN),
		}
	}

	return profiles, nil
}
