import { useEffect, type Dispatch } from "react";
import {
  createInitialNewStoryDialogForm,
  type DeadlineSource,
  type StoryFormAction,
} from "./new-story-dialog-form";

type InitialFormInput = Parameters<typeof createInitialNewStoryDialogForm>[0];

export const useNewStoryDialogInitialization = ({
  assigneeId,
  autoAssignedUserId,
  autoSchedulingDefaultEnabled,
  currentTeamId,
  deadlineSourceRef,
  defaultStatusId,
  dispatch,
  initialDeadlineSource,
  initialSprintEndDate,
  objectiveId,
  priority,
  sprintId,
}: {
  assigneeId: InitialFormInput["assigneeId"];
  autoAssignedUserId: InitialFormInput["autoAssignedUserId"];
  autoSchedulingDefaultEnabled: InitialFormInput["autoSchedulingDefaultEnabled"];
  currentTeamId: InitialFormInput["currentTeamId"];
  deadlineSourceRef: { current: DeadlineSource };
  defaultStatusId: InitialFormInput["statusId"];
  dispatch: Dispatch<StoryFormAction>;
  initialDeadlineSource: DeadlineSource;
  initialSprintEndDate: InitialFormInput["endDate"];
  objectiveId: InitialFormInput["objectiveId"];
  priority: InitialFormInput["priority"];
  sprintId: InitialFormInput["sprintId"];
}) => {
  useEffect(() => {
    dispatch({
      type: "INITIALIZE",
      payload: createInitialNewStoryDialogForm({
        assigneeId,
        autoAssignedUserId,
        autoSchedulingDefaultEnabled,
        currentTeamId,
        endDate: initialSprintEndDate,
        objectiveId,
        priority,
        sprintId,
        statusId: defaultStatusId,
      }),
    });
    deadlineSourceRef.current = initialDeadlineSource;
  }, [
    assigneeId,
    autoAssignedUserId,
    autoSchedulingDefaultEnabled,
    currentTeamId,
    deadlineSourceRef,
    defaultStatusId,
    dispatch,
    initialDeadlineSource,
    initialSprintEndDate,
    objectiveId,
    priority,
    sprintId,
  ]);
};
