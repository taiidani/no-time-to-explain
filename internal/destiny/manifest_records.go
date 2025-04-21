package destiny

import "github.com/taiidani/go-bungie-api/api"

// https://bungie-net.github.io/multi/schema_Destiny-Definitions-Records-DestinyRecordDefinition.html
type RecordDefinition struct {
	DisplayProperties    api.Destiny_Definitions_Common_DestinyDisplayPropertiesDefinition `json:"displayProperties"`
	Scope                int                                                               `json:"scope"`
	LoreHash             int                                                               `json:"loreHash"`
	ObjectiveHashes      []int                                                             `json:"objectiveHashes"`
	RecordValueStyle     int                                                               `json:"recordValueStyle"`
	ForTitleGilding      bool                                                              `json:"forTitleGilding"`
	ShouldShowLargeIcons bool                                                              `json:"shouldShowLargeIcons"`
	TitleInfo            RecordTitleBlock                                                  `json:"titleInfo"`

	CompletionInfo struct {
		PartialCompletionObjectiveCountThreshold int  `json:"partialCompletionObjectiveCountThreshold"`
		ScoreValue                               int  `json:"ScoreValue"`
		ShouldFireToast                          bool `json:"shouldFireToast"`
		ToastStyle                               int  `json:"toastStyle"`
	} `json:"completionInfo"`

	StateInfo struct {
		FeaturedPriority                int    `json:"featuredPriority"`
		ObscuredName                    string `json:"obscuredName"`
		ObscuredDescription             string `json:"obscuredDescription"`
		CompleteUnlockHash              int    `json:"completeUnlockHash"`
		ClaimedUnlockHash               int    `json:"claimedUnlockHash"`
		CompletedCounterUnlockValueHash int    `json:"completedCounterUnlockValueHash"`
	} `json:"stateInfo"`

	Requirements struct {
		EntitlementUnavailableMessage string `json:"entitlementUnavailableMessage"`
	} `json:"requirements"`

	ExpirationInfo struct {
		HasExpiration bool   `json:"hasExpiration"`
		Description   string `json:"description"`
	} `json:"expirationInfo"`

	IntervalInfo api.Destiny_Definitions_Records_DestinyRecordIntervalBlock `json:"intervalInfo"`

	RewardItems                       []api.Destiny_DestinyItemQuantity `json:"rewardItems"`
	AnyRewardHasConditionalVisibility bool                              `json:"anyRewardHasConditionalVisibility"`
	RecordTypeName                    string                            `json:"recordTypeName"`
	PresentationNodeType              int                               `json:"presentationNodeType"`
	TraitIds                          []string                          `json:"traitIds"`
	TraitHashes                       []int                             `json:"traitHashes"`
	ParentNodeHashes                  []int                             `json:"parentNodeHashes"`
	Hash                              int                               `json:"hash"`
	Index                             int                               `json:"index"`
	Redacted                          bool                              `json:"redacted"`
	Blacklisted                       bool                              `json:"blacklisted"`
}

// https://bungie-net.github.io/multi/schema_Destiny-Definitions-Records-DestinyRecordTitleBlock.html
type RecordTitleBlock struct {
	HasTitle bool `json:"hasTitle"`

	TitlesByGender struct {
		Male   string `json:"Male"`
		Female string `json:"Female"`
	} `json:"titlesByGender"`

	TitlesByGenderHash struct {
		Male   string `json:"Male"`
		Female string `json:"Female"`
	} `json:"titlesByGenderHash"`

	GildingTrackingRecordHash uint32 `json:"gildingTrackingRecordHash"`
}
