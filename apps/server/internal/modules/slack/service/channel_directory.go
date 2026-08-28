package slack

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

func (s *Service) fetchChannels(ctx context.Context, botToken string) ([]slackChannelPayload, error) {
	cursor := ""
	channels := make([]slackChannelPayload, 0)
	seenCursors := make(map[string]struct{})

	for {
		endpoint := "https://slack.com/api/conversations.list?limit=200&types=public_channel,private_channel"
		if cursor != "" {
			endpoint += "&cursor=" + url.QueryEscape(cursor)
		}
		var response struct {
			OK       bool   `json:"ok"`
			Error    string `json:"error"`
			Channels []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				IsPrivate  bool   `json:"is_private"`
				IsArchived bool   `json:"is_archived"`
				IsMember   bool   `json:"is_member"`
			} `json:"channels"`
			ResponseMetadata struct {
				NextCursor string `json:"next_cursor"`
			} `json:"response_metadata"`
		}
		if err := s.callSlackAPI(ctx, botToken, endpoint, nil, &response); err != nil {
			return nil, err
		}
		for _, channel := range response.Channels {
			if strings.TrimSpace(channel.ID) == "" {
				continue
			}
			name := strings.TrimSpace(channel.Name)
			if name == "" {
				name = channel.ID
			}
			channels = append(channels, slackChannelPayload{
				SlackChannelID: channel.ID,
				Name:           name,
				IsPrivate:      channel.IsPrivate,
				IsArchived:     channel.IsArchived,
				IsMember:       channel.IsMember,
			})
		}
		nextCursor := strings.TrimSpace(response.ResponseMetadata.NextCursor)
		if nextCursor == "" {
			break
		}
		if _, seen := seenCursors[nextCursor]; seen {
			return nil, errors.New("slack channel pagination repeated a cursor")
		}
		seenCursors[nextCursor] = struct{}{}
		cursor = nextCursor
	}

	return channels, nil
}
