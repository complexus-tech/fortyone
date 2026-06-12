import type { EstimateScheme } from "@/lib/estimate";
import type { NewStory } from "@/modules/story/types";

export const buildNewStoryDialogPayload = ({
  currentTeamId,
  description,
  descriptionHTML,
  estimateScheme,
  storyForm,
  title,
}: {
  currentTeamId?: string;
  description: string;
  descriptionHTML: string;
  estimateScheme: EstimateScheme;
  storyForm: NewStory;
  title: string;
}): NewStory => ({
  title,
  description,
  descriptionHTML,
  teamId: currentTeamId,
  priority: storyForm.priority,
  statusId: storyForm.statusId,
  endDate: storyForm.endDate,
  startDate: storyForm.startDate,
  assigneeId: storyForm.assigneeId,
  objectiveId: storyForm.objectiveId,
  sprintId: storyForm.sprintId,
  estimateScheme,
  estimateValue: storyForm.estimateValue ?? null,
  labelIds: storyForm.labelIds ?? [],
});
