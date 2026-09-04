package workspaces

import (
	"errors"
	"html"

	"github.com/google/uuid"
)

var ErrNoExampleStatuses = errors.New("workspace examples require an unstarted or backlog team status")

type exampleTemplate struct {
	title       string
	description string
}

// BuildExampleStories creates a small, clearly labeled starting point without
// suggesting that work has already started or that a deadline has been agreed.
func BuildExampleStories(teamID, userID uuid.UUID, statuses []SeedStatus, workType WorkType) ([]SeedStory, error) {
	if err := (CreationOptions{WorkType: workType}).Validate(); err != nil {
		return nil, err
	}
	statusID := exampleStatus(statuses)
	if statusID == uuid.Nil {
		return nil, ErrNoExampleStatuses
	}

	templates := exampleTemplates(workType)
	result := make([]SeedStory, len(templates))
	for index, template := range templates {
		description := "This is an example. Edit or delete it to match your work. " + template.description
		descriptionHTML := "<p>" + html.EscapeString(description) + "</p>"
		result[index] = SeedStory{
			Title:       "[Example] " + template.title,
			Description: &description, DescriptionHTML: &descriptionHTML,
			Reporter: userID, Assignee: userID, Team: teamID,
			Priority: "No Priority", Status: statusID,
		}
	}
	return result, nil
}

func exampleStatus(statuses []SeedStatus) uuid.UUID {
	var backlogID uuid.UUID
	for _, status := range statuses {
		if status.ID == uuid.Nil {
			continue
		}
		switch status.Category {
		case "unstarted":
			return status.ID
		case "backlog":
			if backlogID == uuid.Nil {
				backlogID = status.ID
			}
		}
	}
	return backlogID
}

func exampleTemplates(workType WorkType) [3]exampleTemplate {
	switch workType {
	case WorkTypeProduct:
		return [3]exampleTemplate{
			{"Define the next product improvement", "Describe a problem you want to solve and how you will know the improvement helps."},
			{"Build a small prototype", "Outline the smallest version you can use to test the idea."},
			{"Review feedback and choose the next step", "Capture what you learn from the prototype and decide what to improve next."},
		}
	case WorkTypeMarketing:
		return [3]exampleTemplate{
			{"Plan the next campaign", "Choose an audience, the message you want to share, and an outcome to measure."},
			{"Draft the first campaign message", "Write a first version of the message and choose where you would share it."},
			{"Review campaign results", "Record the results when they are available and choose what to try next."},
		}
	case WorkTypeOperations:
		return [3]exampleTemplate{
			{"Document a recurring process", "Write down the steps of a process you repeat and the outcome each step supports."},
			{"Improve one handoff", "Identify information that gets lost between steps and suggest a clearer handoff."},
			{"Review the weekly workflow", "Look for repeated delays or extra work and choose one improvement to try."},
		}
	case WorkTypePersonal:
		return [3]exampleTemplate{
			{"Choose one goal for this week", "Describe a personal outcome that matters to you and what finishing it would look like."},
			{"Break the goal into a small next step", "Choose one manageable action that moves you toward your goal."},
			{"Review what worked", "Reflect on what helped, what got in the way, and what you want to try next."},
		}
	default:
		return [3]exampleTemplate{
			{"Define the outcome you want", "Describe what you want to accomplish and how you will recognize progress."},
			{"Choose the first small task", "Write down one clear action you can take toward that outcome."},
			{"Review progress and decide what comes next", "Capture what you learn and choose the next useful action."},
		}
	}
}
