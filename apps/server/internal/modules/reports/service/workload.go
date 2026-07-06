package reports

import "sort"

const (
	overloadedOpenStoryThreshold = 8
	overloadedEstimateThreshold  = 20
	workloadRiskLimit            = 10
)

func deriveWorkloadRisks(analysis CoreWorkloadAnalysis) CoreWorkloadRisks {
	overloadedMembers := make([]CoreMemberWorkload, 0)
	overdueMembers := make([]CoreMemberWorkload, 0)

	for _, member := range analysis.Members {
		if member.OpenStories >= overloadedOpenStoryThreshold || member.EstimateTotal >= overloadedEstimateThreshold {
			overloadedMembers = append(overloadedMembers, member)
		}
		if member.OverdueStories > 0 {
			overdueMembers = append(overdueMembers, member)
		}
	}

	sort.SliceStable(overloadedMembers, func(i, j int) bool {
		if overloadedMembers[i].EstimateTotal == overloadedMembers[j].EstimateTotal {
			return overloadedMembers[i].OpenStories > overloadedMembers[j].OpenStories
		}
		return overloadedMembers[i].EstimateTotal > overloadedMembers[j].EstimateTotal
	})
	sort.SliceStable(overdueMembers, func(i, j int) bool {
		if overdueMembers[i].OverdueStories == overdueMembers[j].OverdueStories {
			return overdueMembers[i].EstimateTotal > overdueMembers[j].EstimateTotal
		}
		return overdueMembers[i].OverdueStories > overdueMembers[j].OverdueStories
	})

	return CoreWorkloadRisks{
		OverloadedMembers:   limitWorkloadMembers(overloadedMembers),
		OverdueMembers:      limitWorkloadMembers(overdueMembers),
		UnassignedStories:   analysis.Summary.UnassignedStories,
		UnestimatedStories:  analysis.Summary.UnestimatedStories,
		HighPriorityStories: analysis.Summary.HighPriorityStories + analysis.Summary.UrgentStories,
	}
}

func limitWorkloadMembers(members []CoreMemberWorkload) []CoreMemberWorkload {
	if len(members) <= workloadRiskLimit {
		return members
	}

	return members[:workloadRiskLimit]
}
