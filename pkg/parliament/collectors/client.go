package collectors

import (
	"context"
	"strconv"

	"github.com/xaphere/parlament-scripts/pkg/parliament/models"
)

func GetMembers(ctx context.Context) ([]*models.Member, error) {
	reader, err := GetPageReader(ctx, parliamentMPAddress)
	if err != nil {
		return nil, err
	}
	memberIDs, err := ExtractMemberIDs(reader)
	if err != nil {
		return nil, err
	}

	members := []*models.Member{}
	for _, id := range memberIDs {
		memberID, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		reader, err = GetPageReader(ctx, parliamentMPAddress+id)
		if err != nil {
			return nil, err
		}
		member, err := ExtractMember(reader, memberID, GetLocalPartyID, GetLocalConstituencyID)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}
