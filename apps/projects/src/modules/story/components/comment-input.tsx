import { useEffect, useRef } from "react";
import { useEditor } from "@tiptap/react";
import { toast } from "sonner";
import { Button, Flex, Text, TextEditor } from "ui";
import { AttachmentIcon } from "icons";
import { cn } from "lib";
import { DesktopDraftNotice } from "@/shell/desktop/desktop-draft-notice";
import { useDesktopDraft } from "@/shell/desktop/use-desktop-draft";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import { useBrowserCommentDraft } from "@/modules/story/hooks/use-browser-comment-draft";
import type { CommentDraft } from "@/modules/story/hooks/use-browser-comment-draft";
import {
  COMMENT_ATTACHMENT_ACCEPT,
  useCommentAttachments,
} from "@/modules/story/hooks/use-comment-attachments";
import { useCommentMentions } from "../hooks/use-comment-mentions";
import {
  useCommentSubmission,
  useCommentSubmissionIdentity,
} from "../hooks/use-comment-submission";
import { setCommentAttachmentHandler } from "./comment-attachment-extension";
import { CommentUploadStatus } from "./comment-upload-status";
import { getStoryCommentEditorExtensions } from "./story-comment-editor";

type CommentInputProps = {
  storyId: string;
  parentId?: string;
  commentId?: string;
  className?: string;
  initialComment?: string;
  teamId: string;
  onCancel?: () => void;
};

const CommentComposer = ({
  storyId,
  parentId,
  commentId,
  className,
  initialComment,
  onCancel,
  teamId,
}: CommentInputProps) => {
  const attachmentInput = useRef<HTMLInputElement>(null);
  const attachments = useCommentAttachments(storyId);
  let draftKey = `comment:${storyId}`;
  if (parentId) draftKey += `:reply:${parentId}`;
  if (commentId) draftKey = `comment:${storyId}:edit:${commentId}`;
  const draft = useDesktopDraft<CommentDraft>(draftKey);
  const browserDraft = useBrowserCommentDraft(draftKey, !draft.supported);
  const submission = useCommentSubmissionIdentity();
  const activeDraft = draft.supported ? draft : browserDraft;
  const ready = activeDraft.ready && !activeDraft.recovered;

  const getPlaceHolder = () => {
    if (commentId) {
      return "Edit comment...";
    }
    if (parentId) {
      return "Reply to comment...";
    }
    return "Leave a comment...";
  };

  const getButtonLabel = () => {
    if (commentId) {
      return "Update";
    }
    if (parentId) {
      return "Reply";
    }
    return "Comment";
  };

  const editor = useEditor({
    extensions: getStoryCommentEditorExtensions({
      enableMentions: true,
      enableAttachments: true,
      placeholder: getPlaceHolder(),
    }),
    content: initialComment ?? "",
    editable: ready,
    onUpdate: ({ editor: current }) => {
      if (!activeDraft.ready || activeDraft.recovered) return;
      const html = current.getHTML();
      if (submission.get()?.html !== html) submission.set(null);
      const identity = submission.get();
      void activeDraft
        .save({
          version: 1,
          html,
          ...(identity ? { creationId: identity.creationId } : {}),
        })
        .catch(() => {
          // Desktop notices and browser toasts expose storage failures.
        });
    },
    immediatelyRender: false,
  });
  const { enabled: mayaEnabled } = useCommentMentions(
    editor,
    storyId,
    teamId,
    !commentId,
  );

  const { handleComment, isPending, canEdit } = useCommentSubmission({
    editor,
    ready,
    attachments,
    mayaEnabled,
    storyId,
    parentId,
    commentId,
    onCancel,
    submission,
    activeDraft,
    draft,
    browserDraft,
  });

  useEffect(() => {
    editor?.setEditable(canEdit, false);
  }, [canEdit, editor]);

  useEffect(() => {
    if (editor) setCommentAttachmentHandler(editor, attachments.addFiles);
  }, [attachments.addFiles, editor]);

  const { recovered: browserRecovered, acceptRecovered } = browserDraft;
  useEffect(() => {
    if (!editor || draft.supported || !browserRecovered) return;
    submission.set(
      browserRecovered.creationId
        ? {
            html: browserRecovered.html,
            creationId: browserRecovered.creationId,
          }
        : null,
    );
    editor.commands.setContent(browserRecovered.html, {
      emitUpdate: false,
      parseOptions: { preserveWhitespace: "full" },
    });
    acceptRecovered();
  }, [acceptRecovered, browserRecovered, draft.supported, editor, submission]);

  return (
    <Flex
      aria-busy={isPending}
      className={cn(
        "border-border/40 bg-surface-muted/40 ml-1 min-h-24 w-full rounded-xl border px-4 pb-4",
        className,
      )}
      direction="column"
      gap={2}
      justify="between"
    >
      <TextEditor
        className="story-comment-content prose-base prose-a:text-foreground @container/comment-composer leading-6 antialiased"
        editor={editor}
      />
      <DesktopDraftNotice
        error={draft.error}
        onDiscard={() => {
          void draft
            .discard()
            .catch(() => toast.error("Could not discard the saved draft."));
        }}
        onRestore={() => {
          if (
            draft.recovered?.version !== 1 ||
            typeof draft.recovered.html !== "string"
          )
            return;
          editor?.commands.setContent(draft.recovered.html, {
            emitUpdate: false,
            parseOptions: { preserveWhitespace: "full" },
          });
          submission.set(
            draft.recovered.creationId
              ? {
                  html: draft.recovered.html,
                  creationId: draft.recovered.creationId,
                }
              : null,
          );
          draft.acceptRecovered();
        }}
        onRetry={() => {
          void draft.retry();
        }}
        recovered={Boolean(draft.recovered)}
        status={draft.status}
      />
      <CommentUploadStatus
        onRemove={attachments.remove}
        onRetry={(id) => {
          if (editor) attachments.retry(editor, id);
        }}
        uploads={attachments.uploads}
      />
      {commentId && initialComment?.includes('data-mention-kind="agent"') ? (
        <Text className="text-xs" color="muted">
          Editing a comment does not send a new request to Maya.
        </Text>
      ) : null}
      <Flex align="center" gap={2} justify="between">
        <input
          accept={COMMENT_ATTACHMENT_ACCEPT}
          aria-label="Attach files to comment"
          className="hidden"
          multiple
          onChange={(event) => {
            if (editor)
              attachments.addFiles(
                editor,
                Array.from(event.target.files ?? []),
              );
            event.target.value = "";
          }}
          ref={attachmentInput}
          type="file"
        />
        <Button
          aria-label="Attach files"
          asIcon
          color="tertiary"
          disabled={!canEdit || !editor}
          onClick={() => attachmentInput.current?.click()}
          size="sm"
          variant="naked"
        >
          <AttachmentIcon className="size-4" />
        </Button>
        <Flex gap={2} justify="end">
          {onCancel ? (
            <Button
              className="px-3"
              color="tertiary"
              disabled={isPending || attachments.isUploading}
              onClick={onCancel}
              size="sm"
              variant="outline"
            >
              Cancel
            </Button>
          ) : null}
          <Button
            className="px-3"
            color="tertiary"
            disabled={!canEdit || !editor || attachments.hasBlockingUploads}
            onClick={handleComment}
            size="sm"
            variant="outline"
          >
            {getButtonLabel()}
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};

export const CommentInput = (props: CommentInputProps) => {
  const scope = useDesktopWorkspaceScope();
  return (
    <CommentComposer
      {...props}
      key={JSON.stringify([
        scope?.accountId,
        scope?.workspaceId,
        props.storyId,
        props.parentId,
        props.commentId,
      ])}
    />
  );
};
