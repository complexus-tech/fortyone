import type { Editor } from "@tiptap/core";
import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { storyKeys } from "@/shared/story/cache-keys";
import { addAttachmentAction } from "../actions/add-attachment";
import {
  FREE_ATTACHMENT_SIZE_LIMIT,
  MAX_ATTACHMENT_BATCH_FILES,
  PAID_ATTACHMENT_SIZE_LIMIT,
  STORY_ATTACHMENT_ACCEPT,
} from "../components/attachment-upload";
import {
  commentAttachmentUrl,
  insertCommentAttachment,
} from "../components/comment-attachments";

export type AttachmentUpload = {
  id: string;
  file: File;
  status: "uploading" | "failed";
  error?: string;
};

export const COMMENT_ATTACHMENT_ACCEPT = Object.entries(STORY_ATTACHMENT_ACCEPT)
  .flatMap(([mimeType, extensions]) => [mimeType, ...extensions])
  .join(",");

export const validateCommentAttachment = (file: File, maximumSize: number) => {
  const extension = `.${file.name.split(".").at(-1)?.toLowerCase()}`;
  const accepted = Object.entries(STORY_ATTACHMENT_ACCEPT).some(
    ([mimeType, extensions]) =>
      mimeType === file.type || extensions.includes(extension),
  );
  if (!accepted)
    throw new Error(
      "Choose an image, MP4, PDF, Office document, text, or CSV file.",
    );
  if (file.size === 0) throw new Error("The file is empty.");
  if (file.size > maximumSize)
    throw new Error(
      `Files cannot be larger than ${maximumSize / 1024 / 1024} MB.`,
    );
};

export const useCommentAttachments = (storyId: string) => {
  const { workspaceSlug } = useWorkspacePath();
  const { getLimit } = useSubscriptionFeatures();
  const queryClient = useQueryClient();
  const [uploads, setUploads] = useState<AttachmentUpload[]>([]);
  const uploadsRef = useRef<AttachmentUpload[]>([]);
  const mounted = useRef(true);
  const maximumSize =
    getLimit("maxFileUploads") === "10MB"
      ? FREE_ATTACHMENT_SIZE_LIMIT
      : PAID_ATTACHMENT_SIZE_LIMIT;

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  const updateUploads = useCallback(
    (update: (current: AttachmentUpload[]) => AttachmentUpload[]) => {
      uploadsRef.current = update(uploadsRef.current);
      if (mounted.current) setUploads(uploadsRef.current);
    },
    [],
  );

  const upload = useCallback(
    async (editor: Editor, entry: AttachmentUpload) => {
      updateUploads((current) =>
        current.map((item) =>
          item.id === entry.id
            ? { ...item, status: "uploading", error: undefined }
            : item,
        ),
      );
      try {
        const response = await addAttachmentAction(
          storyId,
          entry.file,
          workspaceSlug,
        );
        if (response.error || !response.data) {
          throw new Error(
            response.error?.message || "The file could not be uploaded.",
          );
        }
        void queryClient.invalidateQueries({
          queryKey: storyKeys.attachments(workspaceSlug, storyId),
        });
        // The file remains a durable story attachment if the composer closes.
        if (!mounted.current || editor.isDestroyed) return;
        const attachment = {
          ...response.data,
          url: commentAttachmentUrl(workspaceSlug, storyId, response.data.id),
        };
        const inserted = insertCommentAttachment(editor, attachment);
        if (!inserted)
          throw new Error(
            "The file was uploaded but could not be added to this comment.",
          );
        updateUploads((current) =>
          current.filter((item) => item.id !== entry.id),
        );
      } catch (error) {
        updateUploads((current) =>
          current.map((item) =>
            item.id === entry.id
              ? {
                  ...item,
                  status: "failed",
                  error:
                    error instanceof Error
                      ? error.message
                      : "The file could not be uploaded.",
                }
              : item,
          ),
        );
      }
    },
    [queryClient, storyId, updateUploads, workspaceSlug],
  );

  const addFiles = useCallback(
    (editor: Editor, files: File[]) => {
      if (!editor.isEditable || files.length === 0) return;
      if (files.length > MAX_ATTACHMENT_BATCH_FILES) {
        toast.error(
          `Choose up to ${MAX_ATTACHMENT_BATCH_FILES} files at a time.`,
        );
        return;
      }
      const accepted: AttachmentUpload[] = [];
      for (const file of files) {
        try {
          validateCommentAttachment(file, maximumSize);
          accepted.push({ id: crypto.randomUUID(), file, status: "uploading" });
        } catch (error) {
          toast.error(`Could not attach ${file.name}`, {
            description:
              error instanceof Error ? error.message : "Invalid file.",
          });
        }
      }
      updateUploads((current) => [...current, ...accepted]);
      for (const entry of accepted) void upload(editor, entry);
    },
    [maximumSize, updateUploads, upload],
  );

  const retry = (editor: Editor, id: string) => {
    const entry = uploadsRef.current.find((item) => item.id === id);
    if (entry?.status === "failed") void upload(editor, entry);
  };
  const remove = (id: string) => {
    updateUploads((current) =>
      current.filter((item) => item.id !== id || item.status === "uploading"),
    );
  };

  return {
    addFiles,
    hasBlockingUploads: uploads.length > 0,
    hasPendingUploads: () => uploadsRef.current.length > 0,
    isUploading: uploads.some((item) => item.status === "uploading"),
    remove,
    retry,
    uploads,
  };
};
