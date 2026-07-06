package maya

import (
	"fmt"
	"sort"
	"strings"
)

const (
	minWorkFocusEvidenceStories = 6
	minWorkFocusKeywordScore    = 4
	minWorkFocusConfidence      = 0.45
)

type WorkFocusEvidence struct {
	Title       string
	Description string
	Labels      []string
}

type WorkFocusInferenceInput struct {
	ManualRoleTitle       string
	ManualRoleDescription string
	Evidence              []WorkFocusEvidence
}

type WorkFocusInferenceResult struct {
	ShouldInfer     bool
	RoleTitle       string
	RoleDescription string
	StoryCount      int
	Confidence      float64
}

type workFocusCategory struct {
	Key         string
	RoleTitle   string
	Description string
	Keywords    []string
}

var workFocusCategories = []workFocusCategory{
	{
		Key:         "frontend",
		RoleTitle:   "Frontend engineer",
		Description: "Usually works on frontend and UI work in this team, based on recent issues around interfaces, components, and client-side behavior.",
		Keywords: []string{
			"frontend", "front-end", "react", "next", "ui", "ux", "component", "components",
			"layout", "css", "tailwind", "responsive", "page", "sidebar", "dialog", "modal",
			"menu", "toolbar", "button", "hover", "client", "browser", "form", "card", "cards",
		},
	},
	{
		Key:         "backend",
		RoleTitle:   "Backend engineer",
		Description: "Usually works on backend and platform work in this team, based on recent issues around APIs, data, integrations, and automation.",
		Keywords: []string{
			"backend", "back-end", "api", "endpoint", "server", "database", "db", "sql",
			"migration", "repository", "query", "worker", "job", "queue", "webhook", "auth",
			"oauth", "integration", "redis", "cache", "cron", "scheduler", "service",
		},
	},
	{
		Key:         "product",
		RoleTitle:   "Product engineer",
		Description: "Usually works across product workflow and feature delivery in this team, based on recent issues around user flows, requirements, and product behavior.",
		Keywords: []string{
			"product", "workflow", "flow", "feature", "requirements", "user", "experience",
			"copy", "settings", "onboarding", "analytics", "report", "dashboard", "insight",
		},
	},
	{
		Key:         "design",
		RoleTitle:   "Product designer",
		Description: "Usually works on product design and interface quality in this team, based on recent issues around layout, interaction, and visual polish.",
		Keywords: []string{
			"design", "visual", "polish", "spacing", "typography", "font", "color", "icon",
			"prototype", "figma", "mockup", "accessibility", "empty state", "skeleton",
		},
	},
}

func InferWorkFocus(input WorkFocusInferenceInput) WorkFocusInferenceResult {
	if strings.TrimSpace(input.ManualRoleTitle) != "" || strings.TrimSpace(input.ManualRoleDescription) != "" {
		return WorkFocusInferenceResult{}
	}
	storyCount := len(input.Evidence)
	if storyCount < minWorkFocusEvidenceStories {
		return WorkFocusInferenceResult{StoryCount: storyCount}
	}

	scores := make(map[string]int, len(workFocusCategories))
	totalScore := 0
	for _, evidence := range input.Evidence {
		text := strings.ToLower(strings.Join([]string{
			evidence.Title,
			evidence.Description,
			strings.Join(evidence.Labels, " "),
		}, " "))
		for _, category := range workFocusCategories {
			for _, keyword := range category.Keywords {
				if strings.Contains(text, keyword) {
					scores[category.Key]++
					totalScore++
				}
			}
		}
	}
	if totalScore == 0 {
		return WorkFocusInferenceResult{StoryCount: storyCount}
	}

	categories := append([]workFocusCategory(nil), workFocusCategories...)
	sort.SliceStable(categories, func(i, j int) bool {
		leftScore := scores[categories[i].Key]
		rightScore := scores[categories[j].Key]
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return categories[i].Key < categories[j].Key
	})

	selected := categories[0]
	selectedScore := scores[selected.Key]
	confidence := float64(selectedScore) / float64(totalScore)
	if selectedScore < minWorkFocusKeywordScore || confidence < minWorkFocusConfidence {
		return WorkFocusInferenceResult{
			StoryCount: storyCount,
			Confidence: confidence,
		}
	}

	return WorkFocusInferenceResult{
		ShouldInfer:     true,
		RoleTitle:       selected.RoleTitle,
		RoleDescription: fmt.Sprintf("%s Reviewed %d recent issues.", selected.Description, storyCount),
		StoryCount:      storyCount,
		Confidence:      confidence,
	}
}
