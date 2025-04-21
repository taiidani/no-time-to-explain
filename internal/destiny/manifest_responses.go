package destiny

import "github.com/taiidani/go-bungie-api/api"

// URL: https://bungie-net.github.io/multi/schema_Destiny-Responses-DestinyProfileResponse.html
type ProfileResponse struct {
	// ComponentTypeRecords
	ProfileRecords struct {
		Data ProfileRecordsComponent `json:"data"`
	} `json:"profileRecords"`

	// ComponentTypeRecords
	CharacterRecords struct {
		Data map[string]struct {
			FeaturedRecordHashes []int                              `json:"featuredRecordHashes"`
			Records              map[string]ProfileRecordsComponent `json:"records"`
		} `json:"data"`
	} `json:"characterRecords"`

	// ComponentTypeMetrics
	Metrics struct {
		Data struct {
			Metrics map[string]ProfileMetric `json:"metrics"`
		} `json:"data"`
	} `json:"metrics"`
}

type ProfileMetric struct {
	Invisble          bool                                        `json:"invisible"`
	ObjectiveProgress api.Destiny_Quests_DestinyObjectiveProgress `json:"objectiveProgress"`
}
