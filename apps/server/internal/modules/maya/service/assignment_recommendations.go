package maya

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p Planner) RecommendAssignments(ctx context.Context, input BatchAssignmentRecommendationInput) (BatchAssignmentRecommendationResult, error) {
	input.Candidates = preferRecentlyActiveRecommendations(input.Candidates)
	if p.advisor != nil {
		if batchAdvisor, ok := p.advisor.(BatchAssignmentAdvisor); ok {
			result, err := batchAdvisor.RecommendAssignments(ctx, input)
			if err == nil && len(result.Assignments) > 0 {
				return result, nil
			}
		}
	}
	return deterministicBatchAssignments(input), nil
}

func deterministicBatchAssignments(input BatchAssignmentRecommendationInput) BatchAssignmentRecommendationResult {
	if len(input.Candidates) == 0 || len(input.Stories) == 0 {
		return BatchAssignmentRecommendationResult{}
	}
	candidates := append([]CandidateRecommendation(nil), input.Candidates...)
	sort.SliceStable(candidates, func(i, j int) bool {
		left := candidates[i]
		right := candidates[j]
		if left.EstimateTotal != right.EstimateTotal {
			return left.EstimateTotal < right.EstimateTotal
		}
		if left.OpenStories != right.OpenStories {
			return left.OpenStories < right.OpenStories
		}
		return left.FullName < right.FullName
	})
	assignments := make([]BatchAssignmentRecommendation, 0, len(input.Stories))
	for index, story := range input.Stories {
		candidate := candidates[index%len(candidates)]
		assignments = append(assignments, BatchAssignmentRecommendation{
			StoryID:    story.ID,
			AssigneeID: candidate.UserID,
			Reason:     assignmentReasonForCandidate(candidate),
		})
	}
	return BatchAssignmentRecommendationResult{Assignments: assignments}
}

func assignmentReasonForCandidate(candidate CandidateRecommendation) string {
	if strings.TrimSpace(candidate.TeamAIRoleTitle) != "" {
		return fmt.Sprintf("Maya selected %s because their work focus is %s and their current workload is lighter than the alternatives.", displayCandidateName(candidate), candidate.TeamAIRoleTitle)
	}
	return fmt.Sprintf("Maya selected %s based on current workload and availability.", displayCandidateName(candidate))
}

func preferRecentlyActiveChoices(candidates []candidateChoice) []candidateChoice {
	active := make([]candidateChoice, 0, len(candidates))
	for _, candidate := range candidates {
		if isRecentlyActive(candidate.candidate.Member.LastStoryActivityAt) {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return candidates
	}
	return active
}

func preferRecentlyActiveSchedules(candidates []CandidateSchedule) []CandidateSchedule {
	active := make([]CandidateSchedule, 0, len(candidates))
	for _, candidate := range candidates {
		if isRecentlyActive(candidate.Member.LastStoryActivityAt) {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return candidates
	}
	return active
}

func preferRecentlyActiveRecommendations(candidates []CandidateRecommendation) []CandidateRecommendation {
	active := make([]CandidateRecommendation, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.RecentlyActive {
			active = append(active, candidate)
		}
	}
	if len(active) == 0 {
		return candidates
	}
	return active
}

func isRecentlyActive(lastActivityAt *time.Time) bool {
	if lastActivityAt == nil {
		return false
	}
	return time.Since(lastActivityAt.UTC()) <= recentActivityDays*24*time.Hour
}

func daysSinceLastActivity(lastActivityAt *time.Time) *int {
	if lastActivityAt == nil {
		return nil
	}
	days := int(time.Since(lastActivityAt.UTC()).Hours() / 24)
	if days < 0 {
		days = 0
	}
	return &days
}

func displayCandidateName(candidate CandidateRecommendation) string {
	if strings.TrimSpace(candidate.FullName) != "" {
		return candidate.FullName
	}
	if strings.TrimSpace(candidate.Username) != "" {
		return candidate.Username
	}
	return candidate.UserID.String()
}

type candidateChoice struct {
	candidate         CandidateSchedule
	slot              timeSlot
	plan              timeSlot
	segments          []timeSlot
	remainingMinutes  int
	preemptedBlockIDs []uuid.UUID
}

type timeSlot struct {
	start time.Time
	end   time.Time
}

type segmentPlan struct {
	segments         []timeSlot
	remainingMinutes int
}
