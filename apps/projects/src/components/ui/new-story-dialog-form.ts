import type { NewStory } from "@/modules/story/types";

export const buildNewStoryDialogPayload = ({
  currentTeamId,
  description,
  descriptionHTML,
  storyForm,
  title,
}: {
  currentTeamId?: string;
  description: string;
  descriptionHTML: string;
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
  estimateValue: storyForm.estimateValue ?? null,
  labelIds: storyForm.labelIds ?? [],
});
