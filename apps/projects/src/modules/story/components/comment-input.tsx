import { useEditor } from "@tiptap/react";
import { toast } from "sonner";
import { Button, Flex, TextEditor } from "ui";
import { cn } from "lib";
import { useCommentStoryMutation } from "@/modules/story/hooks/comment-mutation";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { Member } from "@/types";
import { extractMentionsFromHTML } from "@/lib/utils/mentions";
import { type MentionItem } from "./mentions/list";
import { getStoryCommentEditorExtensions } from "./story-comment-editor";

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
  const { data: members = [] } = useTeamMembers(teamId);
  const { mutate } = useCommentStoryMutation();
  const { mutate: updateComment } = useUpdateCommentMutation();
  const mentionUsers: MentionItem[] = members.map((member: Member) => ({
    id: member.id,
    label: member.fullName || member.username,
    username: member.username,
    avatar: member.avatarUrl,
  }));

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
      mentionUsers,
      placeholder: getPlaceHolder(),
    }),
    content: initialComment ?? "",
    editable: true,
    immediatelyRender: false,
  });

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
          onSuccess: () => {
            editor?.commands.clearContent();
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
          onSuccess: () => {
            editor?.commands.clearContent();
            onCancel?.();
          },
        },
      );
    }
  };

  return (
    <Flex
      className={cn(
        "border-border/40 bg-surface-muted/40 ml-1 min-h-24 w-full rounded-2xl border px-4 pb-4",
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
