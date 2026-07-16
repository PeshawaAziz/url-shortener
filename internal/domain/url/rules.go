package url

type RequestContext struct {
	IPAddress   string
	UserAgent   string
	CountryCode string
	DeviceType  string
}

type RoutingConfig struct {
	Variants    []Variant    `json:"variants,omitempty"`
	GeoRules    []GeoRule    `json:"geo_rules,omitempty"`
	DeviceRules []DeviceRule `json:"device_rules,omitempty"`
}

type Variant struct {
	Destination string `json:"destination"`
	Weight      int    `json:"weight"`
}

type GeoRule struct {
	CountryCode string `json:"country_code"`
	Destination string `json:"destination"`
}

type DeviceRule struct {
	DeviceType  string `json:"device_type"`
	Destination string `json:"destination"`
}
