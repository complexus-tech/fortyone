import { useEffect } from "react";
import { useEditor } from "@tiptap/react";
import { toast } from "sonner";
import { Button, Flex, TextEditor } from "ui";
import { cn } from "lib";
import { DesktopDraftNotice } from "@/shell/desktop/desktop-draft-notice";
import { useDesktopDraft } from "@/shell/desktop/use-desktop-draft";
import { useCommentStoryMutation } from "@/modules/story/hooks/comment-mutation";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { extractMentionsFromHTML } from "@/lib/utils/mentions";
import { type MentionItem } from "./mentions/list";
import {
  getStoryCommentEditorExtensions,
  setStoryCommentMentionUsers,
} from "./story-comment-editor";

export const CommentInput = ({
  storyId,
  parentId,
  commentId,
  className,
  initialComment,
  onCancel,
  teamId,
}: {
  storyId: string;
  parentId?: string;
  commentId?: string;
  className?: string;
  initialComment?: string;
  teamId: string;
  onCancel?: () => void;
}) => {
  const { data: members } = useTeamMembers(teamId);
  let draftKey = `comment:${storyId}`;
  if (parentId) draftKey += `:reply:${parentId}`;
  if (commentId) draftKey = `comment:${storyId}:edit:${commentId}`;
  const draft = useDesktopDraft<{ version: 1; html: string }>(draftKey);
  const { mutate } = useCommentStoryMutation();
  const { mutate: updateComment } = useUpdateCommentMutation();

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
    editable: !draft.supported || (draft.ready && !draft.recovered),
    onUpdate: ({ editor: current }) => {
      if (!draft.supported || !draft.ready || draft.recovered) return;
      void draft.save({ version: 1, html: current.getHTML() }).catch(() => {
        // The draft notice keeps the local storage failure visible.
      });
    },
    immediatelyRender: false,
  });

  useEffect(() => {
    editor?.setEditable(
      !draft.supported || (draft.ready && !draft.recovered),
      false,
    );
  }, [draft.ready, draft.recovered, draft.supported, editor]);

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

  const handleComment = () => {
    const comment = editor?.getHTML() ?? "";
    if (editor?.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }

    const mentions = extractMentionsFromHTML(comment).map((m) => m.id);

    if (commentId) {
      // update comment
      updateComment(
        {
          commentId,
          payload: { content: comment, mentions },
          storyId,
        },
        {
          onSuccess: async () => {
            editor?.commands.clearContent();
            try {
              await draft.discard();
            } catch {
              toast.error(
                "Comment saved, but the local draft could not be removed.",
              );
            }
            onCancel?.();
          },
        },
      );
    } else {
      mutate(
        {
          storyId,
          payload: {
            comment,
            parentId: parentId ?? null,
            mentions,
          },
        },
        {
          onSuccess: async () => {
            editor?.commands.clearContent();
            try {
              await draft.discard();
            } catch {
              toast.error(
                "Comment saved, but the local draft could not be removed.",
              );
            }
            onCancel?.();
          },
        },
      );
    }
  };

  return (
    <Flex
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
          disabled={
            draft.supported ? !draft.ready || Boolean(draft.recovered) : null
          }
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
