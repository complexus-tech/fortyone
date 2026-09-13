import type { RichTextValue } from "@/components/rich-text/content";
import type { StoryPriority } from "@/modules/stories/types";
import type { FormAction, SheetName } from "../form-state";
import { useState } from "react";
import * as Crypto from "expo-crypto";
import { useRouter } from "expo-router";
import { useDraft } from "@/components/rich-text/use-draft";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useTeamMembers } from "@/modules/members/hooks/use-team-members";
import { useTeamLabels } from "@/modules/labels/hooks/use-labels";
import { formReducer, initialState, isFormState } from "../form-state";
import { useCreateStoryMutation } from "./use-create-story-mutation";

/** Owns the draft lifecycle and submission; the screen only presents form values. */
export const useNewStoryForm = () => {
  const router = useRouter();
  const mutation = useCreateStoryMutation();
  const [initialDraft] = useState(() => ({
    ...initialState,
    idempotencyKey: Crypto.randomUUID(),
  }));
  const draft = useDraft("new-story", initialDraft, isFormState);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const state = draft.value;

  const { data: teams = [] } = useTeams();
  const selectedTeamId = state.teamId || teams[0]?.id || "";
  const { data: statuses = [] } = useTeamStatuses(selectedTeamId);
  const { data: members = [] } = useTeamMembers(selectedTeamId);
  const { data: labels = [] } = useTeamLabels(selectedTeamId);

  const dispatch = (action: FormAction) => {
    if (!draft.ready || mutation.isPending) return;
    draft.update(formReducer(draft.valueRef.current, action));
  };

  const canSubmit =
    draft.ready && state.title.trim().length > 0 && Boolean(selectedTeamId);

  const submit = () => {
    if (!canSubmit || mutation.isPending) return;
    setSubmitError(null);
    mutation.mutate(
      {
        title: state.title.trim(),
        description: state.description.text || undefined,
        descriptionHTML: state.description.html || undefined,
        idempotencyKey: state.idempotencyKey,
        teamId: selectedTeamId,
        statusId: state.statusId,
        assigneeId: state.assigneeId,
        priority: state.priority,
        labelIds: state.labelIds,
      },
      {
        onSuccess: async (response) => {
          if (response.data?.id) {
            try {
              await draft.clear();
              router.replace(`/story/${response.data.id}`);
            } catch {
              setSubmitError(
                "Your story was created, but this device could not remove its draft. Try again to reopen the same story safely.",
              );
            }
          }
        },
        onError: (cause) =>
          setSubmitError(
            cause instanceof Error
              ? cause.message
              : "Your task could not be saved. Your draft is still here.",
          ),
      },
    );
  };

  const selectMetadata = (id: string) => {
    switch (state.activeSheet) {
      case "team":
        dispatch({ type: "setTeam", teamId: id });
        break;
      case "status":
        dispatch({ type: "setStatus", statusId: id });
        break;
      case "priority":
        dispatch({ type: "setPriority", priority: id as StoryPriority });
        break;
      case "assignee":
        dispatch({ type: "setAssignee", assigneeId: id });
        break;
      case "labels":
        dispatch({ type: "toggleLabel", labelId: id });
        break;
    }
  };

  const setDescription = (description: RichTextValue) =>
    draft.persist(
      formReducer(draft.valueRef.current, {
        type: "setDescription",
        description,
      }),
    );

  return {
    state,
    draft,
    teams,
    statuses,
    members,
    labels,
    selectedTeamId,
    selectedTeam: teams.find((team) => team.id === selectedTeamId),
    selectedStatus: statuses.find((status) => status.id === state.statusId),
    selectedAssignee: members.find((member) => member.id === state.assigneeId),
    selectedLabels: labels.filter((label) => state.labelIds.includes(label.id)),
    canSubmit,
    isSubmitting: mutation.isPending,
    submitError,
    submit,
    selectMetadata,
    setTitle: (title: string) => dispatch({ type: "setTitle", title }),
    setDescription,
    openSheet: (sheet: SheetName) => dispatch({ type: "setSheet", sheet }),
    closeSheet: () => dispatch({ type: "setSheet", sheet: null }),
  };
};
