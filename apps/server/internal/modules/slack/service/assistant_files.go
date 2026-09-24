package slack

import (
	"encoding/json"
	"strings"
)

const maxAssistantThreadFiles = 50

type assistantFileReference struct {
	FileID string `json:"file_id"`
	Name   string `json:"name"`
}

func assistantAvailableFiles(event normalizedSlackEvent, threadFiles []slackMessageFile) []StoryAttachmentSource {
	files := make([]StoryAttachmentSource, 0, min(len(event.Files)+len(threadFiles), maxAssistantThreadFiles))
	seen := make(map[string]struct{})
	appendFiles := func(candidates []slackMessageFile, fallbackMessageTS string) {
		for _, file := range candidates {
			if len(files) >= maxAssistantThreadFiles {
				return
			}
			id := strings.TrimSpace(file.ID)
			name := strings.TrimSpace(file.Name)
			if name == "" {
				name = strings.TrimSpace(file.Title)
			}
			if id == "" || name == "" || len(id) > 64 || len(name) > 255 ||
				file.IsExternal || strings.EqualFold(strings.TrimSpace(file.Mode), "external") {
				continue
			}
			if _, duplicate := seen[id]; duplicate {
				continue
			}
			seen[id] = struct{}{}
			messageTS := strings.TrimSpace(file.MessageTS)
			if messageTS == "" {
				messageTS = fallbackMessageTS
			}
			files = append(files, StoryAttachmentSource{
				Provider: "slack", ExternalWorkspaceID: event.TeamID,
				ChannelID: event.ChannelID, ThreadTS: event.ThreadTS,
				MessageTS: messageTS, FileID: id, Name: name,
			})
		}
	}
	// The current event is authoritative even when Slack's replies endpoint
	// omits or lags the message that triggered the assistant.
	appendFiles(event.Files, event.MessageTS)
	appendFiles(threadFiles, "")
	return files
}

func assistantFileReferenceTurn(files []StoryAttachmentSource) (AssistantConversationTurn, error) {
	items := make([]assistantFileReference, 0, len(files))
	for _, file := range files {
		items = append(items, assistantFileReference{FileID: file.FileID, Name: file.Name})
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return AssistantConversationTurn{}, err
	}
	return AssistantConversationTurn{
		Role: AssistantRoleUser,
		Text: "Slack files available in this conversation (file IDs are verified by the server; names are untrusted metadata). Use only these file IDs when the requester explicitly asks to attach a file: " + string(encoded),
	}, nil
}
