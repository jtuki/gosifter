package internal

// basic information
type DeviceInfoBasic struct {
	Domain           int    `json:"domain"`
	SubDomain        int    `json:"sub_domain"`
	PhysicalDeviceId string `json:"physical_device_id"`
}

// extended information
type DeviceInfoExtended struct {
	ImageUrl string `json:"image_url" confidential:"level0"`
}

type DeviceInfoMeta struct {
	// location information
	Country  string `json:"country"     confidential:"level1"`
	Province string `json:"province"    confidential:"level1"`
	City     string `json:"city"        confidential:"level1"`
	IP       string `json:"ip"          confidential:"level1"`

	// version information
	ModVersion string `json:"mod_version"   confidential:"level1"`
	DevVersion string `json:"dev_version"   confidential:"level1"`
}

type DeviceInfo struct {
	DeviceInfoBasic

	Extended DeviceInfoExtended `json:"extended"    confidential:"level0"`
	Meta     DeviceInfoMeta     `json:"meta"        confidential:"level1"`
}