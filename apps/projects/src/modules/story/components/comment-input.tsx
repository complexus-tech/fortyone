import { useEditor, ReactRenderer, mergeAttributes } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Mention from "@tiptap/extension-mention";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import tippy from "tippy.js";
import { toast } from "sonner";
import { Button, Flex, TextEditor } from "ui";
import { cn } from "lib";
import { useCommentStoryMutation } from "@/modules/story/hooks/comment-mutation";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { Member } from "@/types";
import { extractMentionsFromHTML } from "@/lib/utils/mentions";
import {
  MentionList,
  type MentionItem,
  type MentionListRef,
} from "./mentions/list";

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
    extensions: [
      StarterKit,
      Underline,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        autolink: true,
      }),
      Mention.configure({
        HTMLAttributes: {
          class:
            "mention bg-surface-muted hover:bg-state-hover transition",
        },
        renderHTML({ options, node }) {
          return [
            "a",
            mergeAttributes(
              {
                href: `/profile/${node.attrs.id}`,
              },
              options.HTMLAttributes,
            ),
            `${options.suggestion.char}${node.attrs.label ?? node.attrs.id}`,
          ];
        },
        suggestion: {
          items: ({ query }: { query: string }) => {
            if (!query || query.trim() === "") {
              return mentionUsers.slice(0, 6);
            }

            const filtered = mentionUsers.filter(
              (user) =>
                user.label.toLowerCase().includes(query.toLowerCase()) ||
                user.username.toLowerCase().includes(query.toLowerCase()),
            );

            return filtered.slice(0, 6);
          },
          render: () => {
            let component: ReactRenderer<MentionListRef>;
            // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tippy.js instance type is complex
            let popup: any;

            return {
              // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
              onStart: (props: any) => {
                component = new ReactRenderer(MentionList, {
                  props,
                  editor: props.editor,
                });

                if (!props.clientRect) {
                  return;
                }

                popup = tippy("body", {
                  getReferenceClientRect: props.clientRect,
                  appendTo: () => document.body,
                  content: component.element,
                  showOnCreate: true,
                  interactive: true,
                  trigger: "manual",
                  placement: "bottom-start",
                });
              },

              // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
              onUpdate(props: any) {
                component.updateProps(props as Record<string, unknown>);
                if (!props.clientRect) {
                  return;
                }

                popup?.[0]?.setProps({
                  getReferenceClientRect: props.clientRect,
                });
              },

              // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
              onKeyDown(props: any) {
                if (props.event.key === "Escape") {
                  popup?.[0]?.hide();
                  return true;
                }

                return (
                  component.ref?.onKeyDown(props.event as KeyboardEvent) ??
                  false
                );
              },

              onExit() {
                popup?.[0]?.destroy();
                component.destroy();
              },
            };
          },
        },
      }),
      Placeholder.configure({
        placeholder: getPlaceHolder(),
      }),
    ],
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
        "ml-1 min-h-24 w-full rounded-2xl border border-border/40 bg-surface-muted/40 bg-surface/50 dark:shadow-dark-300/50",
        className,
      )}
      direction="column"
      gap={2}
      justify="between"
    >
      <TextEditor
        className="prose-base leading-6 antialiased prose-a:text-foreground"
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
