package destiny

import "github.com/taiidani/go-bungie-api/api"

// URL: https://bungie-net.github.io/multi/schema_Destiny-Config-GearAssetDataBaseDefinition.html
type Manifest struct {
	Version                  string                                           `json:"version"`
	MobileAssetContentPath   string                                           `json:"mobileAssetContentPath"`
	MobileGearAssetDataBases []api.Destiny_Config_GearAssetDataBaseDefinition `json:"mobileGearAssetDataBases"`
	MobileWorldContentPaths  struct {
		English string `json:"en"`
	} `json:"mobileWorldContentPaths"`
	JsonWorldContentPaths struct {
		English string `json:"en"`
	} `json:"jsonWorldContentPaths"`
	JsonWorldComponentContentPaths struct {
		English map[string]string `json:"en"`
	} `json:"jsonWorldComponentContentPaths"`
	MobileClanBannerDatabasePath string `json:"mobileClanBannerDatabasePath"`
	MobileGearCDN                struct {
		Geometry    string
		Texture     string
		PlateRegion string
		Gear        string
		Shader      string
	} `json:"mobileGearCDN"`
	IconImagePyramidInfo []api.Destiny_Config_ImagePyramidEntry `json:"iconImagePyramidInfo"`
}
