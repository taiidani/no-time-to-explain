package destiny

import (
	"context"
	"fmt"
	"log/slog"
)

type HelperClanTitles struct {
	TotalScore int
	Member     []HelperClanTitlesMember
}

type HelperClanTitlesMember struct {
	Name   string
	Titles []HelperClanTitle
}

type HelperClanTitle struct {
	Name        string
	Icon        string
	Earned      bool
	GildedCount int
}

// GetClanTitles tracks the triumph and title progress for the clan.
func (h *Helper) GetClanTitles(ctx context.Context) (*HelperClanTitles, error) {
	manifest, err := h.client.GetManifest(ctx)
	if err != nil {
		return nil, err
	}

	// Figure out all of the title manifest entries
	recordsManifestURL := manifest.JsonWorldComponentContentPaths.English["DestinyRecordDefinition"]
	recordsManifest, err := h.client.GetRecordManifestDefinition(ctx, recordsManifestURL)
	if err != nil {
		return nil, err
	}

	titles := map[string]RecordDefinition{}
	for key, entry := range recordsManifest {
		if !entry.TitleInfo.HasTitle {
			continue
		}

		titles[key] = entry
	}

	members, err := h.client.GetClanMembers(ctx, unknownSpaceGroupID)
	if err != nil {
		return nil, err
	}

	ret := &HelperClanTitles{}
	for _, member := range members {
		profile, err := h.client.GetProfile(ctx, member.DestinyUserInfo.MembershipType, member.DestinyUserInfo.MembershipId)
		if err != nil {
			return nil, err
		}

		clanMember := HelperClanTitlesMember{
			Name:   member.DestinyUserInfo.DisplayName,
			Titles: []HelperClanTitle{},
		}

		for key, record := range profile.ProfileRecords.Data.Records {
			manifestDefinition, ok := titles[key]
			if !ok {
				slog.Warn("Failed to associate manifest entry with key", "key", key)
				continue
			}

			// Check for gildings
			gildedCount := 0
			if manifestDefinition.TitleInfo.GildingTrackingRecordHash > 0 {
				gildedRecordKey := fmt.Sprintf("%d", manifestDefinition.TitleInfo.GildingTrackingRecordHash)
				gildedRecord, ok := profile.ProfileRecords.Data.Records[gildedRecordKey]
				if !ok {
					slog.Warn("Failed to look up gilding record for manifest entry", "record-key", gildedRecordKey, "manifest-key", key)
					continue
				}

				gildedCount = int(*gildedRecord.CompletedCount)
			}

			// Record the title
			title := HelperClanTitle{
				Name:        manifestDefinition.DisplayProperties.Name,
				Icon:        manifestDefinition.DisplayProperties.Icon,
				Earned:      record.Objectives[0].Complete,
				GildedCount: gildedCount,
			}

			clanMember.Titles = append(clanMember.Titles, title)
		}

		ret.Member = append(ret.Member, clanMember)
	}

	return ret, nil
}
