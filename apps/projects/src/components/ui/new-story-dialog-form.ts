import type { DetailedStory, NewStory } from "@/modules/story/types";
import { isMayaAssigneeSelection } from "@/lib/auto-scheduling";
import { normalizeTimeNeeded } from "@/lib/time-needed";

export type DeadlineSource = "unset" | "sprint" | "manual" | "cleared";

export const getInitialDeadlineSource = ({
  sprintId,
  sprintEndDate,
}: {
  sprintId?: string | null;
  sprintEndDate?: string | null;
}): DeadlineSource => (sprintId && sprintEndDate ? "sprint" : "unset");

export const getDeadlineForSprintSelection = ({
  currentEndDate,
  currentSource,
  sprintEndDate,
}: {
  currentEndDate: string | null | undefined;
  currentSource: DeadlineSource;
  sprintEndDate?: string | null;
}): { endDate: string | null; source: DeadlineSource } => {
  const shouldInheritSprintDeadline =
    currentSource === "unset" || currentSource === "sprint";

  if (!shouldInheritSprintDeadline) {
    return {
      endDate: currentEndDate ?? null,
      source: currentSource,
    };
  }

  return sprintEndDate
    ? { endDate: sprintEndDate, source: "sprint" }
    : { endDate: null, source: "unset" };
};

export const buildNewStoryDialogPayload = ({
  currentTeamId,
  description,
  descriptionHTML,
  mayaAssigneeId,
  storyForm,
  title,
}: {
  currentTeamId?: string;
  description: string;
  descriptionHTML: string;
  mayaAssigneeId?: string | null;
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
    autoSchedulingEnabled: isMayaAssigneeSelection(
      storyForm.assigneeId,
      mayaAssigneeId,
    )
      ? true
      : storyForm.autoSchedulingEnabled ?? true,
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
