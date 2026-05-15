package vote

import (
	"errors"
	"testing"

	"github.com/OpenSlides/openslides-go/datastore/dsfetch"
	"github.com/OpenSlides/openslides-go/datastore/dsmock"
	"github.com/OpenSlides/openslides-go/datastore/dsmodels"
)

func TestAllowedToVoteWithMultipleDelegations(t *testing.T) {
	ctx := t.Context()

	baseData := `---
	meeting/1:
		users_enable_vote_delegations: true
		users_forbid_delegator_to_vote: false

	user:
		10:
			is_present_in_meeting_ids: [1]
		20:
			is_present_in_meeting_ids: [1]
		30:
			is_present_in_meeting_ids: [1]

	meeting_user:
		100:
			user_id: 10
			meeting_id: 1
			group_ids: [5]
		200:
			user_id: 20
			meeting_id: 1
			group_ids: [5]
		300:
			user_id: 30
			meeting_id: 1
			group_ids: [6]
`

	for _, tt := range []struct {
		name                     string
		data                     string
		representedMeetingUserID int
		actingMeetingUserID      int
		expectAllowed            bool
	}{
		{
			name: "acting user is one of multiple delegates",
			data: `---
			meeting_user/100/vote_delegated_to_ids: [300, 200]
			`,
			representedMeetingUserID: 100,
			actingMeetingUserID:      200,
			expectAllowed:            true,
		},
		{
			name: "acting user is not delegated",
			data: `---
			meeting_user/100/vote_delegated_to_ids: [300]
			`,
			representedMeetingUserID: 100,
			actingMeetingUserID:      200,
			expectAllowed:            false,
		},
		{
			name: "delegator may vote for self when not forbidden",
			data: `---
			meeting_user/100/vote_delegated_to_ids: [200]
			`,
			representedMeetingUserID: 100,
			actingMeetingUserID:      100,
			expectAllowed:            true,
		},
		{
			name: "delegator may not vote for self when forbidden",
			data: `---
			meeting/1/users_forbid_delegator_to_vote: true
			meeting_user/100/vote_delegated_to_ids: [200]
			`,
			representedMeetingUserID: 100,
			actingMeetingUserID:      100,
			expectAllowed:            false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			data := dsmock.YAMLData(baseData)
			for key, value := range dsmock.YAMLData(tt.data) {
				data[key] = value
			}

			fetch := dsfetch.New(dsmock.Stub(data))
			poll := dsmodels.Poll{
				ID:               1,
				MeetingID:        1,
				EntitledGroupIDs: []int{5},
			}

			err := allowedToVote(ctx, fetch, poll, tt.representedMeetingUserID, tt.actingMeetingUserID)
			if tt.expectAllowed {
				if err != nil {
					t.Fatalf("allowedToVote returned unexpected error: %v", err)
				}
				return
			}

			if !errors.Is(err, ErrNotAllowed) {
				t.Fatalf("expected ErrNotAllowed, got %v", err)
			}
		})
	}
}
