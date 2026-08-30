import type { DetailedStory, NewStory } from "@/modules/story/types";
import { isMayaAssigneeSelection } from "@/lib/auto-scheduling";
import {
  DEFAULT_TIME_NEEDED_MINUTES,
  normalizeTimeNeeded,
} from "@/lib/time-needed";

export type DeadlineSource = "unset" | "sprint" | "manual" | "cleared";
export type NewStoryDialogForm = NewStory;

export type StoryFormAction =
  | { type: "INITIALIZE"; payload: NewStory }
  | {
      type: "SET_FIELD";
      field: keyof NewStory;
      value: NewStory[keyof NewStory];
    }
  | { type: "RESET_FORM"; payload: NewStory }
  | { type: "PATCH_FORM"; payload: Partial<NewStory> }
  | { type: "SYNC_TEAM_STATUS"; teamId: string; statusId: string };

export const storyFormReducer = (
  state: NewStory,
  action: StoryFormAction,
): NewStory => {
  switch (action.type) {
    case "INITIALIZE":
      return action.payload;
    case "SET_FIELD":
      return { ...state, [action.field]: action.value };
    case "PATCH_FORM":
      return { ...state, ...action.payload };
    case "SYNC_TEAM_STATUS":
      return {
        ...state,
        teamId: action.teamId,
        statusId: action.statusId,
      };
    case "RESET_FORM":
      return action.payload;
    default:
      return state;
  }
};

export const toDateOnly = (value?: string | null): string | null => {
  if (!value) return null;

  const dateOnly = /^\d{4}-\d{2}-\d{2}/.exec(value)?.[0];
  return dateOnly ?? value;
};

export const getInitialDeadlineSource = ({
  sprintId,
  sprintEndDate,
}: {
  sprintId?: string | null;
  sprintEndDate?: string | null;
}): DeadlineSource => (sprintId && sprintEndDate ? "sprint" : "unset");

export const createInitialNewStoryDialogForm = ({
  assigneeId,
  autoAssignedUserId,
  autoSchedulingDefaultEnabled,
  currentTeamId,
  endDate,
  objectiveId,
  priority,
  sprintId,
  statusId,
}: {
  assigneeId?: string | null;
  autoAssignedUserId: string | null;
  autoSchedulingDefaultEnabled: boolean;
  currentTeamId?: string;
  endDate: string | null;
  objectiveId?: string;
  priority: NewStory["priority"];
  sprintId?: string;
  statusId?: string;
}): NewStory => ({
  title: "",
  description: "",
  descriptionHTML: "",
  teamId: currentTeamId,
  statusId,
  endDate,
  startDate: null,
  assigneeId: assigneeId !== undefined ? assigneeId : autoAssignedUserId,
  priority,
  objectiveId: objectiveId || null,
  keyResultId: null,
  sprintId: sprintId || null,
  estimateValue: null,
  estimatedDurationMinutes: DEFAULT_TIME_NEEDED_MINUTES,
  minimumFocusBlockMinutes: null,
  autoSchedulingEnabled: autoSchedulingDefaultEnabled,
  autoSchedulingLocked: false,
  labelIds: [],
});

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
    endDate: toDateOnly(storyForm.endDate),
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
