import { experimental_useObject as useObject } from "@ai-sdk/react";
import { useCallback, useMemo, useState } from "react";
import { substoryGenerationSchema } from "@/modules/stories/public/substory-generation";
import { useCreateStoryMutation } from "../hooks/create-mutation";
import { isSubstorySuggestionReadyToCreate } from "./substory-suggestion-state";

type UseSubstorySuggestionsArgs = {
  defaultStatusId?: string;
  storyId: string;
  teamId: string;
  workspaceSlug: string;
};

/**
 * Owns the streamed suggestion lifecycle separately from the surrounding
 * substory section. The section composes the trigger and presentation, while
 * this hook keeps request, validation, selection, and creation transitions
 * together.
 */
export const useSubstorySuggestions = ({
  defaultStatusId,
  storyId,
  teamId,
  workspaceSlug,
}: UseSubstorySuggestionsArgs) => {
  const { mutate: createStory, isPending: isCreatingSuggestedSubstories } =
    useCreateStoryMutation();
  const [manualSelectedSubstories, setManualSelectedSubstories] = useState<
    Set<string>
  >(new Set());
  const [hasCustomSelection, setHasCustomSelection] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(true);
  const { object, submit, isLoading, error } = useObject({
    api: "/api/suggest-substories",
    schema: substoryGenerationSchema,
  });

  const suggestedTitles = useMemo(
    () =>
      (object?.substories ?? [])
        .map((substory) => substory?.title)
        .filter((title): title is string => Boolean(title)),
    [object?.substories],
  );
  const selectedSubstories = useMemo(() => {
    if (hasCustomSelection) return manualSelectedSubstories;

    return new Set(suggestedTitles);
  }, [hasCustomSelection, manualSelectedSubstories, suggestedTitles]);
  const canCreateSuggestedSubstories = isSubstorySuggestionReadyToCreate({
    error,
    isLoading,
    object,
  });

  const clearSelection = useCallback(() => {
    setHasCustomSelection(true);
    setManualSelectedSubstories(new Set());
  }, []);

  const requestSuggestions = useCallback(() => {
    setHasCustomSelection(false);
    setManualSelectedSubstories(new Set());
    setShowSuggestions(true);
    submit({ storyId, workspaceSlug });
  }, [storyId, submit, workspaceSlug]);

  const toggleSelectedSubstory = useCallback(
    (substoryTitle: string) => {
      setHasCustomSelection(true);
      setManualSelectedSubstories((previousSelection) => {
        const nextSelection = new Set(
          hasCustomSelection ? previousSelection : suggestedTitles,
        );

        if (nextSelection.has(substoryTitle)) {
          nextSelection.delete(substoryTitle);
        } else {
          nextSelection.add(substoryTitle);
        }

        return nextSelection;
      });
    },
    [hasCustomSelection, suggestedTitles],
  );

  const createSelectedSubstories = useCallback(() => {
    if (!canCreateSuggestedSubstories || isCreatingSuggestedSubstories) {
      return;
    }

    for (const title of selectedSubstories) {
      createStory({
        parentId: storyId,
        priority: "No Priority",
        statusId: defaultStatusId,
        teamId,
        title,
      });
    }

    clearSelection();
    setShowSuggestions(false);
  }, [
    canCreateSuggestedSubstories,
    clearSelection,
    createStory,
    defaultStatusId,
    isCreatingSuggestedSubstories,
    selectedSubstories,
    storyId,
    teamId,
  ]);

  const cancelSuggestions = useCallback(() => {
    clearSelection();
    setShowSuggestions(false);
  }, [clearSelection]);

  return {
    canCreateSuggestedSubstories,
    cancelSuggestions,
    createSelectedSubstories,
    isCreatingSuggestedSubstories,
    isLoadingSuggestions: isLoading,
    isShowingSuggestionError: showSuggestions && Boolean(error),
    requestSuggestions,
    selectedSubstories,
    showSuggestions,
    suggestedSubstories: object?.substories,
    toggleSelectedSubstory,
  };
};
