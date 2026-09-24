package slack

import "strings"

const slackModalMaxFiles = 10

// Slack message shortcuts include the original message, including its file
// objects. Keep only Slack-hosted files: external files are links to a different
// provider and cannot be imported with a Slack bot token.
func sourceFilesFromShortcut(files []slackPayloadFile) []slackSourceFile {
	selected := make([]slackSourceFile, 0, min(len(files), slackModalMaxFiles))
	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		id := strings.TrimSpace(file.ID)
		if !isSlackModalFileID(id) || file.IsExternal || strings.EqualFold(file.Mode, "external") {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		name := strings.TrimSpace(file.Name)
		if name == "" {
			name = strings.TrimSpace(file.Title)
		}
		if name == "" {
			name = id
		}
		selected = append(selected, slackSourceFile{ID: id, Name: truncateRunes(name, slackOptionTextMaxRunes)})
		if len(selected) == slackModalMaxFiles {
			break
		}
	}
	return selected
}

func isSlackModalFileID(id string) bool {
	if len(id) < 2 || len(id) > 64 || id[0] != 'F' {
		return false
	}
	for _, character := range id[1:] {
		if (character < 'A' || character > 'Z') &&
			(character < 'a' || character > 'z') &&
			(character < '0' || character > '9') {
			return false
		}
	}
	return true
}

func uniqueSlackModalFileIDs(ids ...[]string) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, group := range ids {
		for _, rawID := range group {
			id := strings.TrimSpace(rawID)
			if !isSlackModalFileID(id) {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func sourceFileInputBlock(files []slackSourceFile, selectedIDs []string) map[string]any {
	options := make([]map[string]any, 0, len(files))
	initialOptions := make([]map[string]any, 0, len(files))
	selected := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = struct{}{}
	}
	for _, file := range files {
		option := toSlackOption(file.Name, file.ID)
		options = append(options, option)
		if selectedIDs == nil {
			initialOptions = append(initialOptions, option)
		} else if _, exists := selected[file.ID]; exists {
			initialOptions = append(initialOptions, option)
		}
	}
	element := map[string]any{
		"type":      "checkboxes",
		"action_id": modalActionSourceFilesSelect,
		"options":   options,
	}
	if len(initialOptions) > 0 {
		element["initial_options"] = initialOptions
	}
	return map[string]any{
		"type":     "input",
		"block_id": modalBlockSourceFiles,
		"label": map[string]string{
			"type": "plain_text",
			"text": "Files from this message",
		},
		"optional": true,
		"element":  element,
	}
}

func uploadFileInputBlock() map[string]any {
	return map[string]any{
		"type":     "input",
		"block_id": modalBlockUploadFiles,
		"label": map[string]string{
			"type": "plain_text",
			"text": "Upload files",
		},
		"optional": true,
		"element": map[string]any{
			"type":      "file_input",
			"action_id": modalActionUploadFilesInput,
			"max_files": slackModalMaxFiles,
		},
	}
}
