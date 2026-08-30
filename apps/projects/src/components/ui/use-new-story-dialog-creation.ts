import { useState } from "react";
import { toast } from "sonner";
import {
  clearRichTextContent,
  getPersistableRichTextContent,
} from "@/lib/tiptap/rich-text-media";
import { attachFigmaDesigns } from "./new-story-dialog-figma";
import {
  buildNewStoryDialogPayload,
  runStoryCreatedFollowUp,
  type NewStoryDialogForm,
} from "./new-story-dialog-form";

type DescriptionEditor = Parameters<typeof getPersistableRichTextContent>[0];

type TitleEditor = {
  commands: {
    focus: () => unknown;
    setContent: (content: string) => unknown;
  };
  getText: () => string;
};

type CreatedStory = Parameters<typeof runStoryCreatedFollowUp>[0];

type FigmaArtifactReference = {
  canonicalUrl: string;
};

export const useNewStoryDialogCreation = <
  TStory extends CreatedStory,
  TArtifact extends FigmaArtifactReference,
>({
  createMore,
  currentTeamId,
  editor,
  figmaArtifacts,
  finalizeStagedMedia,
  isMayaAssigneeLoading,
  linkFigmaStory,
  mayaAssigneeId,
  mutateStory,
  onCreated,
  onDialogClose,
  onFormReset,
  onFreeStoryCreated,
  onFigmaArtifactsReset,
  onResetDeadlineSource,
  resetStagedMedia,
  setStoryTitle,
  storyForm,
  storyTerm,
  titleEditor,
}: {
  createMore: boolean;
  currentTeamId?: string;
  editor: DescriptionEditor | null;
  figmaArtifacts: TArtifact[];
  finalizeStagedMedia: (
    createdStoryId: string,
    editor: DescriptionEditor,
  ) => Promise<ReturnType<typeof getPersistableRichTextContent> | null>;
  isMayaAssigneeLoading: boolean;
  linkFigmaStory: (input: { storyId: string; url: string }) => Promise<unknown>;
  mayaAssigneeId?: string | null;
  mutateStory: (
    story: ReturnType<typeof buildNewStoryDialogPayload>,
  ) => Promise<TStory>;
  onCreated?: Parameters<typeof runStoryCreatedFollowUp>[1];
  onDialogClose: () => void;
  onFormReset: () => void;
  onFreeStoryCreated?: () => void;
  onFigmaArtifactsReset: () => void;
  onResetDeadlineSource: () => void;
  resetStagedMedia: () => void;
  setStoryTitle: (title: string) => void;
  storyForm: NewStoryDialogForm;
  storyTerm: string;
  titleEditor: TitleEditor | null;
}) => {
  const [isCreating, setIsCreating] = useState(false);

  const handleCreateStory = async () => {
    if (!titleEditor || !editor || isMayaAssigneeLoading) return;

    const title = titleEditor.getText();
    if (!title) {
      titleEditor.commands.focus();
      toast.warning("Validation Error", {
        description: "Title is required",
      });
      return;
    }

    setIsCreating(true);
    const selectedFigmaArtifacts = figmaArtifacts;

    try {
      const initialContent = getPersistableRichTextContent(editor);
      const createdStory = await mutateStory(
        buildNewStoryDialogPayload({
          currentTeamId,
          description: initialContent.contentText,
          descriptionHTML: initialContent.contentHtml,
          mayaAssigneeId,
          storyForm,
          title,
        }),
      );
      let finalizedContent: Awaited<ReturnType<typeof finalizeStagedMedia>> =
        null;

      try {
        finalizedContent = await finalizeStagedMedia(createdStory.id, editor);
      } catch (error) {
        toast.error(`${storyTerm} created, but media could not be finalized`, {
          description:
            error instanceof Error
              ? error.message
              : "The description media could not be saved.",
        });
      }

      const committedStory = finalizedContent
        ? {
            ...createdStory,
            description: finalizedContent.contentText,
            descriptionHTML: finalizedContent.contentHtml,
          }
        : createdStory;

      if (!createMore) onDialogClose();
      titleEditor.commands.setContent("");
      setStoryTitle("");
      clearRichTextContent(editor);
      resetStagedMedia();
      onFormReset();
      onFigmaArtifactsReset();
      onResetDeadlineSource();
      onFreeStoryCreated?.();

      const [followUpError] = await Promise.all([
        runStoryCreatedFollowUp(committedStory, onCreated),
        selectedFigmaArtifacts.length > 0
          ? attachFigmaDesigns({
              artifacts: selectedFigmaArtifacts,
              linkDesign: linkFigmaStory,
              storyId: committedStory.id,
            })
          : Promise.resolve(),
      ]);
      if (followUpError) {
        toast.error(`${storyTerm} created, but the follow-up action failed`, {
          description: followUpError.message,
        });
      }
    } finally {
      setIsCreating(false);
    }
  };

  return { handleCreateStory, isCreating };
};
