package destiny

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/taiidani/no-time-to-explain/internal/models"
)

type HelperClan struct {
	GroupID int
	Members []models.Player
}

func (h *Helper) GetClan(ctx context.Context, groupID int) (*HelperClan, error) {
	members, err := h.client.GetClanMembers(ctx, UnknownSpaceGroupID)
	if err != nil {
		return nil, fmt.Errorf("unable to get clan members: %w", err)
	}

	ret := HelperClan{
		GroupID: groupID,
		Members: []models.Player{},
	}

	for _, member := range members {
		add := models.Player{
			DisplayName:       member.DestinyUserInfo.DisplayName,
			MembershipType:    member.DestinyUserInfo.MembershipType,
			GlobalDisplayName: member.DestinyUserInfo.BungieGlobalDisplayName,
			GlobalDisplayCode: member.DestinyUserInfo.BungieGlobalDisplayNameCode,
			GroupId:           fmt.Sprintf("%d", groupID),
			MembershipId:      member.DestinyUserInfo.MembershipID,
		}

		lastOnlineInt, err := strconv.ParseInt(member.LastOnlineStatusChange, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse last online timestamp %q for %q", member.LastOnlineStatusChange, add.MembershipId)
		}
		add.LastOnline = time.Unix(lastOnlineInt, 0)

		add.GroupJoinDate, err = time.Parse(time.RFC3339, member.JoinDate)
		if err != nil {
			return nil, fmt.Errorf("could not parse join date timestamp %q for %q", member.JoinDate, add.MembershipId)
		}

		ret.Members = append(ret.Members, add)
	}

	return &ret, nil
}
