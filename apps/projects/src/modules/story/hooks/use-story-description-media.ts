"use client";

import { useCallback, useEffect, useRef } from "react";
import type { Editor } from "@tiptap/core";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useTerminology, useWorkspacePath } from "@/hooks";
import {
  getPersistableRichTextContent,
  uploadRichTextMediaFiles,
} from "@/lib/tiptap/rich-text-media";
import { storyKeys } from "@/modules/stories/constants";
import type { DetailedStory } from "../types";
import {
  deleteStoryMediaAction,
  uploadStoryMediaAction,
} from "../actions/story-media";
import { updateStoryAction } from "../actions/update-story";

type StoryIdWaiter = {
  reject: (error: Error) => void;
  resolve: (storyId: string) => void;
};

type PersistedRichTextContent = ReturnType<
  typeof getPersistableRichTextContent
>;

type StoryMediaQueryClient = Pick<
  ReturnType<typeof useQueryClient>,
  "invalidateQueries" | "setQueryData"
>;

export const syncFinalizedStoryMediaCache = (
  queryClient: StoryMediaQueryClient,
  workspaceSlug: string,
  storyId: string,
  content: PersistedRichTextContent,
) => {
  queryClient.setQueryData<DetailedStory>(
    storyKeys.detail(workspaceSlug, storyId),
    (story) =>
      story
        ? {
            ...story,
            description: content.contentText,
            descriptionHTML: content.contentHtml,
          }
        : story,
  );

  void queryClient.invalidateQueries({
    queryKey: storyKeys.all(workspaceSlug),
  });
};

const waitForPendingUploads = async (pendingUploads: Set<Promise<void>>) => {
  const currentUploads = [...pendingUploads];
  if (currentUploads.length === 0) return;

  await Promise.allSettled(currentUploads);
  return waitForPendingUploads(pendingUploads);
};

export const useStoryDescriptionMedia = (storyId?: string | null) => {
  const { getTermDisplay } = useTerminology();
  const { workspaceSlug } = useWorkspacePath();
  const queryClient = useQueryClient();
  const inputRef = useRef<HTMLInputElement>(null);
  const activeStoryIdRef = useRef(storyId ?? null);
  const storyIdWaitersRef = useRef(new Set<StoryIdWaiter>());
  const pendingUploadsRef = useRef(new Set<Promise<void>>());
  const uploadedMediaStoryIdsRef = useRef(new Map<string, string>());

  const releaseStoryId = useCallback((nextStoryId: string) => {
    activeStoryIdRef.current = nextStoryId;
    storyIdWaitersRef.current.forEach(({ resolve }) => {
      resolve(nextStoryId);
    });
    storyIdWaitersRef.current.clear();
  }, []);

  const waitForStoryId = useCallback(() => {
    const activeStoryId = activeStoryIdRef.current;
    if (activeStoryId) return Promise.resolve(activeStoryId);

    return new Promise<string>((resolve, reject) => {
      storyIdWaitersRef.current.add({ reject, resolve });
    });
  }, []);

  const cancelStagedUploads = useCallback(() => {
    const error = new Error("Story media upload was cancelled.");
    storyIdWaitersRef.current.forEach(({ reject }) => {
      reject(error);
    });
    storyIdWaitersRef.current.clear();
    activeStoryIdRef.current = storyId ?? null;
  }, [storyId]);

  useEffect(() => {
    if (storyId) releaseStoryId(storyId);
  }, [releaseStoryId, storyId]);

  useEffect(
    () => () => {
      const error = new Error(
        "Story media upload was cancelled because the editor closed.",
      );
      storyIdWaitersRef.current.forEach(({ reject }) => {
        reject(error);
      });
      storyIdWaitersRef.current.clear();
    },
    [],
  );

  const openMediaPicker = useCallback(() => {
    inputRef.current?.click();
  }, []);

  const handleMediaFiles = useCallback(
    (editor: Editor, files: File[], position?: number) => {
      const uploadTask = uploadRichTextMediaFiles({
        cleanup: async (media) => {
          const mediaStoryId =
            uploadedMediaStoryIdsRef.current.get(media.id) ??
            activeStoryIdRef.current;
          if (!mediaStoryId) return;

          const response = await deleteStoryMediaAction(
            mediaStoryId,
            media.id,
            workspaceSlug,
          );
          if (response.error) {
            throw new Error(
              response.error.message || "Could not clean up uploaded media.",
            );
          }
          uploadedMediaStoryIdsRef.current.delete(media.id);
        },
        editor,
        files,
        onError: (file, error) => {
          if (
            error instanceof Error &&
            error.message.toLowerCase().includes("cancelled")
          ) {
            return;
          }
          toast.error(`Could not add ${file.name}`, {
            description:
              error instanceof Error
                ? error.message
                : "Could not upload this media file.",
          });
        },
        position,
        upload: async (file) => {
          const mediaStoryId = await waitForStoryId();
          const response = await uploadStoryMediaAction(
            mediaStoryId,
            file,
            workspaceSlug,
          );
          if (response.error || !response.data) {
            throw new Error(
              response.error?.message || "Could not upload this media file.",
            );
          }
          uploadedMediaStoryIdsRef.current.set(response.data.id, mediaStoryId);
          return response.data;
        },
      });

      pendingUploadsRef.current.add(uploadTask);
      void uploadTask.then(
        () => {
          pendingUploadsRef.current.delete(uploadTask);
        },
        () => {
          pendingUploadsRef.current.delete(uploadTask);
        },
      );
    },
    [waitForStoryId, workspaceSlug],
  );

  const finalizeStagedMedia = useCallback(
    async (
      createdStoryId: string,
      editor: Editor,
    ): Promise<PersistedRichTextContent | null> => {
      const hadStagedMedia = pendingUploadsRef.current.size > 0;
      if (!hadStagedMedia) return null;

      releaseStoryId(createdStoryId);
      await waitForPendingUploads(pendingUploadsRef.current);

      const content = getPersistableRichTextContent(editor);
      const response = await updateStoryAction(
        createdStoryId,
        {
          description: content.contentText,
          descriptionHTML: content.contentHtml,
          reconcileDescriptionMedia: true,
        },
        workspaceSlug,
      );
      if (!response.error) {
        syncFinalizedStoryMediaCache(
          queryClient,
          workspaceSlug,
          createdStoryId,
          content,
        );
        return content;
      }

      const uploadedIds = [...uploadedMediaStoryIdsRef.current.entries()]
        .filter(([, mediaStoryId]) => mediaStoryId === createdStoryId)
        .map(([attachmentId]) => attachmentId);
      await Promise.allSettled(
        uploadedIds.map((attachmentId) =>
          deleteStoryMediaAction(createdStoryId, attachmentId, workspaceSlug),
        ),
      );
      uploadedIds.forEach((attachmentId) => {
        uploadedMediaStoryIdsRef.current.delete(attachmentId);
      });
      toast.error(
        `${getTermDisplay("storyTerm", { capitalize: true })} created without media`,
        {
          description:
            response.error.message ||
            "The media upload finished, but the description could not be updated.",
        },
      );
      return null;
    },
    [getTermDisplay, queryClient, releaseStoryId, workspaceSlug],
  );

  const resetForNextStory = useCallback(() => {
    cancelStagedUploads();
    activeStoryIdRef.current = storyId ?? null;
    uploadedMediaStoryIdsRef.current.clear();
  }, [cancelStagedUploads, storyId]);

  return {
    cancelStagedUploads,
    finalizeStagedMedia,
    handleMediaFiles,
    inputRef,
    openMediaPicker,
    resetForNextStory,
  };
};
