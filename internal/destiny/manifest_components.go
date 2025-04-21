package destiny

import "github.com/taiidani/go-bungie-api/api"

// URL: https://bungie-net.github.io/multi/schema_Destiny-Components-Records-DestinyProfileRecordsComponent.html
type ProfileRecordsComponent struct {
	Score         int                                                              `json:"score"`
	ActiveScore   int                                                              `json:"activeScore"`
	LegacyScore   int                                                              `json:"legacyScore"`
	LifetimeScore int                                                              `json:"lifetimeScore"`
	Records       map[string]api.Destiny_Components_Records_DestinyRecordComponent `json:"records"`
}
