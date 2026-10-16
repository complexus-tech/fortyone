import { useEffect, useRef, useState } from "react";
import { useEditor } from "@tiptap/react";
import { toast } from "sonner";
import { Button, Flex, TextEditor } from "ui";
import { cn } from "lib";
import { DesktopDraftNotice } from "@/shell/desktop/desktop-draft-notice";
import { useDesktopDraft } from "@/shell/desktop/use-desktop-draft";
import { useDesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import { useCommentStoryMutation } from "@/modules/story/hooks/comment-mutation";
import { useBrowserCommentDraft } from "@/modules/story/hooks/use-browser-comment-draft";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { extractMentionsFromHTML } from "@/lib/utils/mentions";
import { type MentionItem } from "./mentions/list";
import {
  getStoryCommentEditorExtensions,
  setStoryCommentMentionUsers,
} from "./story-comment-editor";

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
  const { data: members } = useTeamMembers(teamId);
  let draftKey = `comment:${storyId}`;
  if (parentId) draftKey += `:reply:${parentId}`;
  if (commentId) draftKey = `comment:${storyId}:edit:${commentId}`;
  const draft = useDesktopDraft<{ version: 1; html: string }>(draftKey);
  const browserDraft = useBrowserCommentDraft(draftKey, !draft.supported);
  const { mutateAsync: createComment, isPending: isCreating } =
    useCommentStoryMutation();
  const { mutateAsync: updateComment, isPending: isUpdating } =
    useUpdateCommentMutation();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const submissionInFlight = useRef(false);
  const isPending = isSubmitting || isCreating || isUpdating;
  const activeDraft = draft.supported ? draft : browserDraft;
  const canEdit = activeDraft.ready && !activeDraft.recovered && !isPending;

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
      placeholder: getPlaceHolder(),
    }),
    content: initialComment ?? "",
    editable: canEdit,
    onUpdate: ({ editor: current }) => {
      if (!activeDraft.ready || activeDraft.recovered) return;
      void activeDraft
        .save({ version: 1, html: current.getHTML() })
        .catch(() => {
          // Desktop notices and browser toasts expose storage failures.
        });
    },
    immediatelyRender: false,
  });

  useEffect(() => {
    editor?.setEditable(canEdit, false);
  }, [canEdit, editor]);

  const { recovered: browserRecovered, acceptRecovered } = browserDraft;
  useEffect(() => {
    if (!editor || draft.supported || !browserRecovered) return;
    editor.commands.setContent(browserRecovered.html, { emitUpdate: false });
    acceptRecovered();
  }, [acceptRecovered, browserRecovered, draft.supported, editor]);

  useEffect(() => {
    if (!editor) {
      return;
    }

    const mentionUsers: MentionItem[] = (members ?? []).map((member) => ({
      id: member.id,
      label: member.fullName || member.username,
      username: member.username,
      avatar: member.avatarUrl,
    }));
    setStoryCommentMentionUsers(editor, mentionUsers);
  }, [editor, members]);

  const handleComment = async () => {
    if (!editor || !canEdit || submissionInFlight.current) return;
    if (editor.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }

    const comment = editor.getHTML();
    const mentions = extractMentionsFromHTML(comment).map((m) => m.id);
    submissionInFlight.current = true;
    setIsSubmitting(true);
    editor.setEditable(false, false);

    try {
      if (commentId) {
        await updateComment({
          commentId,
          payload: { content: comment, mentions },
          storyId,
        });
      } else {
        await createComment({
          storyId,
          payload: {
            comment,
            parentId: parentId ?? null,
            mentions,
          },
        });
      }
    } catch {
      // Mutation hooks report the failure and retain the editor for retry.
      submissionInFlight.current = false;
      setIsSubmitting(false);
      if (!editor.isDestroyed) editor.setEditable(true, false);
      return;
    }

    try {
      if (draft.supported) {
        await draft.discard(
          (payload) =>
            typeof payload === "object" &&
            payload !== null &&
            "html" in payload &&
            payload.html === comment,
        );
      } else {
        await browserDraft.discard(comment);
      }
    } catch {
      toast.error("Comment saved, but the local draft could not be removed.");
    }
    // Do not erase a newer draft if this composer was reopened or its content
    // changed while the request was in flight.
    if (!editor.isDestroyed && editor.getHTML() === comment) {
      editor.commands.clearContent(false);
      onCancel?.();
    }
    submissionInFlight.current = false;
    setIsSubmitting(false);
    if (!editor.isDestroyed) editor.setEditable(true, false);
  };

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
        className="prose-base prose-a:text-foreground leading-6 antialiased"
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
          editor?.commands.setContent(draft.recovered.html);
          draft.acceptRecovered();
        }}
        onRetry={() => {
          void draft.retry();
        }}
        recovered={Boolean(draft.recovered)}
        status={draft.status}
      />
      <Flex gap={2} justify="end">
        {onCancel ? (
          <Button
            className="px-3"
            color="tertiary"
            disabled={isPending}
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
          disabled={!canEdit || !editor}
          onClick={handleComment}
          size="sm"
          variant="outline"
        >
          {getButtonLabel()}
        </Button>
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
