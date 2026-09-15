import type { RichTextValue } from "@/components/rich-text/content";
import type { StoryPriority } from "@/modules/stories/types";
import type { FormAction, SheetName } from "../form-state";
import type { StorySubmissionState } from "../submission";
import { useEffect, useRef, useState } from "react";
import * as Crypto from "expo-crypto";
import { useDraft } from "@/components/rich-text/use-draft";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useTeamMembers } from "@/modules/members/hooks/use-team-members";
import { useTeamLabels } from "@/modules/labels/hooks/use-labels";
import { useAuthStore } from "@/store/auth";
import { formReducer, initialState, isFormState } from "../form-state";
import { createStoryFinalizer } from "../submission";
import { useCreateStoryMutation } from "./use-create-story-mutation";

/** Owns durable submission; the screen coordinates editor flush and navigation. */
export const useNewStoryForm = () => {
  const mutation = useCreateStoryMutation();
  const [initialDraft] = useState(() => ({
    ...initialState,
    idempotencyKey: Crypto.randomUUID(),
  }));
  const [scope] = useState(() => {
    const { userId, workspace, sessionEpoch } = useAuthStore.getState();
    return { userId, workspace, sessionEpoch };
  });
  const draft = useDraft("new-story", initialDraft, isFormState);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submissionState, setSubmissionState] = useState<StorySubmissionState>({
    isSubmitting: false,
    isFinalized: false,
  });
  const [finalizer] = useState(() => createStoryFinalizer(setSubmissionState));
  const mounted = useRef(true);
  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  const state = draft.value;

  const { data: teams = [] } = useTeams();
  const selectedTeamId = state.teamId || teams[0]?.id || "";
  const { data: statuses = [] } = useTeamStatuses(selectedTeamId);
  const { data: members = [] } = useTeamMembers(selectedTeamId);
  const { data: labels = [] } = useTeamLabels(selectedTeamId);

  const assertActive = () => {
    const current = useAuthStore.getState();
    if (
      !mounted.current ||
      !current.isAuthenticated ||
      current.isLoading ||
      current.userId !== scope.userId ||
      current.workspace !== scope.workspace ||
      current.sessionEpoch !== scope.sessionEpoch
    ) {
      throw new Error("Your session changed. Reopen the composer to continue.");
    }
  };
  const canEditDraft = () => mounted.current && finalizer.canEdit();

  const dispatch = (action: FormAction) => {
    if (!draft.ready || !canEditDraft()) return;
    draft.update(formReducer(draft.valueRef.current, action));
  };

  const canSubmit =
    draft.ready &&
    (submissionState.isFinalized ||
      (state.title.trim().length > 0 &&
        teams.some((team) => team.id === selectedTeamId)));

  const submit = async (description: RichTextValue): Promise<string> => {
    if (!draft.ready)
      throw new Error("Wait for your draft to finish restoring.");
    setSubmitError(null);
    try {
      return await finalizer.submit({
        getDraft: () => draft.valueRef.current,
        description,
        availableTeamIds: teams.map((team) => team.id),
        createIdempotencyKey: Crypto.randomUUID,
        persistDraft: draft.persist,
        createStory: async (snapshot) => {
          const response = await mutation.mutateAsync({
            title: snapshot.title,
            description: snapshot.description.text || undefined,
            descriptionHTML: snapshot.description.html || undefined,
            idempotencyKey: snapshot.idempotencyKey,
            teamId: snapshot.teamId,
            statusId: snapshot.statusId,
            assigneeId: snapshot.assigneeId,
            priority: snapshot.priority,
            labelIds: snapshot.labelIds,
          });
          return response.data?.id ?? "";
        },
        clearDraft: draft.clear,
        assertActive,
      });
    } catch (cause) {
      if (mounted.current) {
        setSubmitError(
          cause instanceof Error
            ? cause.message
            : "Your task could not be saved. Your draft is still here.",
        );
      }
      throw cause;
    }
  };

  const selectMetadata = (id: string) => {
    switch (draft.valueRef.current.activeSheet) {
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

  const setDescription = (description: RichTextValue): Promise<void> => {
    if (!draft.ready || !canEditDraft()) {
      return Promise.reject(
        new Error("This draft is not available for editing."),
      );
    }
    return finalizer.persistEdit(() => {
      assertActive();
      return draft.persist(
        formReducer(draft.valueRef.current, {
          type: "setDescription",
          description,
        }),
      );
    });
  };

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
    ...submissionState,
    canEditDraft,
    submitError,
    submit,
    selectMetadata,
    setTitle: (title: string) => dispatch({ type: "setTitle", title }),
    setDescription,
    openSheet: (sheet: SheetName) => dispatch({ type: "setSheet", sheet }),
    closeSheet: () => dispatch({ type: "setSheet", sheet: null }),
  };
};
