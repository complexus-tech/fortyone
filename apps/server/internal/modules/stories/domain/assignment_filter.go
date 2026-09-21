package domain

import (
	"sort"
	"strings"
	"unicode"
)

// AssignmentFilterResult distinguishes an unknown name from an ambiguous one
// so assistant surfaces can ask for clarification instead of guessing.
type AssignmentFilterResult struct {
	Stories    []StoryList
	Resolved   *AssignmentActor
	Candidates []AssignmentActor
}

// FilterStoriesAssignedBy resolves a human name or username only against the
// assignment actors present in the supplied, already-authorized story set.
func FilterStoriesAssignedBy(stories []StoryList, query string) AssignmentFilterResult {
	query = normalizeAssignmentActorName(query)
	if query == "" {
		return AssignmentFilterResult{Stories: append([]StoryList(nil), stories...)}
	}

	type scoredActor struct {
		actor AssignmentActor
		score int
	}
	actorsByID := make(map[string]scoredActor)
	bestScore := 0
	for _, story := range stories {
		if story.AssignedBy == nil {
			continue
		}
		score := assignmentActorMatchScore(*story.AssignedBy, query)
		if score == 0 {
			continue
		}
		key := story.AssignedBy.ID.String()
		if existing, exists := actorsByID[key]; !exists || score > existing.score {
			actorsByID[key] = scoredActor{actor: *story.AssignedBy, score: score}
		}
		if score > bestScore {
			bestScore = score
		}
	}

	candidates := make([]AssignmentActor, 0, len(actorsByID))
	for _, candidate := range actorsByID {
		if candidate.score == bestScore {
			candidates = append(candidates, candidate.actor)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		left := assignmentActorDisplayName(candidates[i])
		right := assignmentActorDisplayName(candidates[j])
		if left != right {
			return left < right
		}
		return candidates[i].ID.String() < candidates[j].ID.String()
	})
	if len(candidates) != 1 {
		return AssignmentFilterResult{Candidates: candidates}
	}

	resolved := candidates[0]
	filtered := make([]StoryList, 0, len(stories))
	for _, story := range stories {
		if story.AssignedBy != nil && story.AssignedBy.ID == resolved.ID {
			filtered = append(filtered, story)
		}
	}
	return AssignmentFilterResult{
		Stories:    filtered,
		Resolved:   &resolved,
		Candidates: candidates,
	}
}

func assignmentActorMatchScore(actor AssignmentActor, query string) int {
	username := normalizeAssignmentActorName(actor.Username)
	fullName := normalizeAssignmentActorName(actor.FullName)
	if query == username || query == fullName {
		return 3
	}
	for _, token := range strings.Fields(fullName) {
		if query == token {
			return 2
		}
	}
	if strings.Contains(username, query) || strings.Contains(fullName, query) {
		return 1
	}
	return 0
}

func assignmentActorDisplayName(actor AssignmentActor) string {
	if name := strings.TrimSpace(actor.FullName); name != "" {
		return strings.ToLower(name)
	}
	return strings.ToLower(strings.TrimSpace(actor.Username))
}

func normalizeAssignmentActorName(value string) string {
	return strings.ToLower(strings.Join(strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool {
		return unicode.IsSpace(r)
	}), " "))
}
