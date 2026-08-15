import type { DetailedStory, NewStory } from "@/modules/story/types";
import { normalizeTimeNeeded } from "@/lib/time-needed";

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
}): NewStory => {
  const timeNeeded = normalizeTimeNeeded({
    estimatedDurationMinutes: storyForm.estimatedDurationMinutes,
    minimumFocusBlockMinutes: storyForm.minimumFocusBlockMinutes,
  });

  return {
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
    keyResultId: storyForm.keyResultId,
    sprintId: storyForm.sprintId,
    estimateValue: storyForm.estimateValue ?? null,
    ...timeNeeded,
    labelIds: storyForm.labelIds ?? [],
  };
};

export const runStoryCreatedFollowUp = async (
  story: DetailedStory,
  onCreated?: (createdStory: DetailedStory) => Promise<void> | void,
): Promise<Error | null> => {
  if (!onCreated) return null;

  try {
    await onCreated(story);
    return null;
  } catch (error) {
    return error instanceof Error
      ? error
      : new Error("The requested follow-up action could not be completed.");
  }
};
