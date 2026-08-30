import { isMayaAssigneeSelection } from "@/lib/auto-scheduling";

export const getSelectedNewStoryLabels = <TLabel extends { id: string }>(
  labels: TLabel[],
  selectedLabelIds?: string[] | null,
) => {
  const selectedLabelIdSet = new Set(selectedLabelIds ?? []);

  return labels.filter((label) => selectedLabelIdSet.has(label.id));
};

export const getNewStoryDialogFieldSelections = <
  TObjective extends { id: string; name: string; sequenceId: number },
  TKeyResult extends { id: string; name: string },
  TSprint extends { id: string; name: string },
  TMember extends { id: string },
  TMayaAssignee extends TMember,
>({
  currentTeamCode,
  keyResults,
  mayaAssignee,
  members,
  objectives,
  sprints,
  storyForm,
}: {
  currentTeamCode?: string;
  keyResults: TKeyResult[];
  mayaAssignee?: TMayaAssignee | null;
  members: TMember[];
  objectives: TObjective[];
  sprints: TSprint[];
  storyForm: {
    assigneeId?: string | null;
    keyResultId?: string | null;
    objectiveId?: string | null;
    sprintId?: string | null;
  };
}) => {
  const objective = objectives.find(({ id }) => id === storyForm.objectiveId);
  const keyResult = keyResults.find(({ id }) => id === storyForm.keyResultId);
  const strategyLinkLabel = keyResult
    ? `${currentTeamCode}-${objective?.sequenceId} / ${keyResult.name}`
    : objective?.name;
  const member =
    mayaAssignee?.id === storyForm.assigneeId
      ? mayaAssignee
      : members.find(({ id }) => id === storyForm.assigneeId);

  return {
    isMayaAssigned: isMayaAssigneeSelection(
      storyForm.assigneeId,
      mayaAssignee?.id,
    ),
    member,
    sprint: sprints.find(({ id }) => id === storyForm.sprintId),
    strategyLinkLabel,
  };
};
