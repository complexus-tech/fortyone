import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { CloseIcon, PlusIcon } from "icons";
import { toast } from "sonner";
import { Button, Tooltip } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { ChatDraft } from "./chat-rich-text";
import type { ChatAttachment, VoiceNoteUpload } from "./types";
import { uploadAttachment, deleteAttachment } from "./api";
import { formatFileSize } from "./attachment-url";
import { useChatIdentity } from "./chat-context";

const ACCEPTED_FILES =
  "image/*,audio/mp4,audio/webm,.m4a,.webm,video/mp4,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-powerpoint,application/vnd.openxmlformats-officedocument.presentationml.presentation,text/plain,text/csv";

export const useComposerAttachments = (
  draft: ChatDraft,
  onDraftChange: (value: ChatDraft) => void,
  editing: boolean,
) => {
  const { workspaceSlug } = useWorkspacePath();
  const { conversationId } = useChatIdentity();
  const uploadScope = `${workspaceSlug}:${conversationId}`;
  const [activeUploadScope, setActiveUploadScope] = useState<string | null>(
    null,
  );
  const uploading = activeUploadScope === uploadScope;
  const uploadingRef = useRef(false);
  const generation = useRef(0);
  const latest = useRef({ draft, onDraftChange });
  useLayoutEffect(() => {
    latest.current = { draft, onDraftChange };
  }, [draft, onDraftChange]);
  useEffect(() => {
    generation.current += 1;
    uploadingRef.current = false;
    return () => {
      generation.current += 1;
    };
  }, [conversationId, workspaceSlug]);

  const uploadFiles = async (
    files: File[],
    voiceNote?: VoiceNoteUpload,
  ): Promise<boolean> => {
    if (!conversationId || !files.length || uploadingRef.current) return false;
    if ((latest.current.draft.attachments?.length ?? 0) + files.length > 10) {
      toast.error("You can attach up to 10 files to a message.");
      return false;
    }
    if (files.some((file) => file.size > 25 * 1024 * 1024)) {
      toast.error("Each file must be 25 MB or smaller.");
      return false;
    }
    const uploadGeneration = generation.current;
    uploadingRef.current = true;
    setActiveUploadScope(uploadScope);
    return Promise.resolve()
      .then(() =>
        Promise.allSettled(
          files.map((file) =>
            uploadAttachment(
              { workspaceSlug },
              conversationId,
              file,
              voiceNote,
            ),
          ),
        ),
      )
      .then(async (results) => {
        const uploaded = results.flatMap((result) =>
          result.status === "fulfilled" ? [result.value] : [],
        );
        const current = latest.current;
        const available = Math.max(
          0,
          10 - (current.draft.attachments?.length ?? 0),
        );
        const accepted =
          generation.current === uploadGeneration
            ? uploaded.slice(0, available)
            : [];
        const unused = uploaded.slice(accepted.length);
        if (accepted.length) {
          current.onDraftChange({
            ...current.draft,
            attachments: [...(current.draft.attachments ?? []), ...accepted],
          });
        }
        if (unused.length) {
          // Unattached uploads also expire server-side if cleanup cannot reach the API.
          await Promise.allSettled(
            unused.map((attachment) =>
              deleteAttachment(
                { workspaceSlug },
                conversationId,
                attachment.id,
              ),
            ),
          );
        }
        if (generation.current !== uploadGeneration) return false;
        const failed = results.find((result) => result.status === "rejected");
        if (failed?.status === "rejected") {
          toast.error("Some files could not be uploaded", {
            description:
              failed.reason instanceof Error
                ? failed.reason.message
                : "Please try again.",
          });
        }
        if (unused.length)
          toast.error("Only 10 attachments can be added to a message.");
        return accepted.length === files.length;
      })
      .catch((error: unknown) => {
        if (generation.current === uploadGeneration) {
          toast.error("Could not attach these files", {
            description:
              error instanceof Error ? error.message : "Please try again.",
          });
        }
        return false;
      })
      .finally(() => {
        if (generation.current === uploadGeneration) {
          uploadingRef.current = false;
          setActiveUploadScope(null);
        }
      });
  };
  const removeFile = (id: string) => {
    onDraftChange({
      ...draft,
      attachments: draft.attachments?.filter(
        (attachment) => attachment.id !== id,
      ),
    });
    if (!editing)
      void deleteAttachment({ workspaceSlug }, conversationId, id).catch(
        (error: unknown) => {
          toast.error("Could not remove the uploaded file", {
            description:
              error instanceof Error ? error.message : "Please try again.",
          });
        },
      );
  };
  return { uploading, uploadFiles, removeFile };
};

export const ComposerAttachmentList = ({
  attachments,
  busy,
  onRemove,
}: {
  attachments: ChatAttachment[];
  busy: boolean;
  onRemove: (id: string) => void;
}) => (
  <div
    aria-label="Message attachments"
    className="flex flex-wrap gap-2 px-3 pb-2"
  >
    {attachments.map((attachment) => (
      <span
        className="border-border flex max-w-full items-center gap-2 rounded-md border px-2 py-1"
        key={attachment.id}
      >
        <span className="truncate">{attachment.fileName}</span>
        <span className="text-text-muted shrink-0">
          {formatFileSize(attachment.sizeBytes)}
        </span>
        <button
          aria-label={`Remove ${attachment.fileName}`}
          className="text-text-muted hover:text-foreground rounded p-1"
          disabled={busy}
          onClick={() => {
            onRemove(attachment.id);
          }}
          type="button"
        >
          <CloseIcon className="size-3" />
        </button>
      </span>
    ))}
  </div>
);

export const AttachFileButton = ({
  busy,
  uploading,
  onUpload,
}: {
  busy: boolean;
  uploading: boolean;
  onUpload: (files: File[]) => Promise<boolean>;
}) => {
  const input = useRef<HTMLInputElement>(null);
  const { conversationId } = useChatIdentity();
  return (
    <>
      <input
        accept={ACCEPTED_FILES}
        aria-label="Choose attachments"
        className="hidden"
        disabled={busy}
        multiple
        onChange={(event) => {
          const files = Array.from(event.target.files ?? []);
          event.target.value = "";
          void onUpload(files);
        }}
        ref={input}
        type="file"
      />
      <Tooltip title={uploading ? "Uploading files…" : "Attach a file"}>
        <Button
          aria-label="Attach a file"
          asIcon
          className="bg-state-hover text-text-muted size-7 rounded-full [&_svg]:text-current"
          color="tertiary"
          disabled={busy || !conversationId}
          onClick={() => {
            input.current?.click();
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <PlusIcon className="size-5" />
        </Button>
      </Tooltip>
    </>
  );
};
