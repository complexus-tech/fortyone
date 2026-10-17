import type { Editor } from "@tiptap/core";
import { useCallback } from "react";
import { toast } from "sonner";
import { uploadRichTextMediaFiles } from "@/lib/tiptap/rich-text-media";
import {
  deleteDocumentMediaAction,
  uploadDocumentMediaAction,
} from "./actions";

export function useDocumentMedia(documentId: string, workspaceSlug: string) {
  const handleMediaFiles = useCallback(
    (currentEditor: Editor, files: File[], position?: number) => {
      void uploadRichTextMediaFiles({
        cleanup: async (media) => {
          const response = await deleteDocumentMediaAction(
            documentId,
            media.id,
            workspaceSlug,
          );
          if (response.error) {
            throw new Error(
              response.error.message || "Could not clean up uploaded media.",
            );
          }
        },
        editor: currentEditor,
        files,
        position,
        upload: async (file) => {
          const response = await uploadDocumentMediaAction(
            documentId,
            file,
            workspaceSlug,
          );
          if (response.error || !response.data) {
            throw new Error(
              response.error?.message || "Could not upload this media file.",
            );
          }
          return response.data;
        },
        onError: (_file, error) => {
          toast.error(
            error instanceof Error
              ? error.message
              : "Could not upload this media file.",
          );
        },
      });
    },
    [documentId, workspaceSlug],
  );

  return handleMediaFiles;
}
