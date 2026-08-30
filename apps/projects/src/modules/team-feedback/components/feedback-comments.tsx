"use client";

import { useState } from "react";
import { useEditor } from "@tiptap/react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { toast } from "sonner";
import { ReplyIcon } from "icons";
import { Avatar, Box, Button, Flex, Text, TextEditor, TimeAgo } from "ui";
import { useSession } from "@/lib/auth/client";
import { getCommentEditorExtensions } from "@/lib/tiptap/comment-editor";
import { serializeCommentToGitHubMarkdown } from "@/lib/tiptap/comment-markdown";
import { useCreateTeamFeedbackComment } from "../hooks/use-create-comment";
import type { TeamFeedbackComment } from "../types";
import { getCommentThreads } from "../utils/comment-threads";

export const FeedbackCommentComposer = ({
  feedbackId,
  onCancel,
  onSubmitted,
  parentId,
}: {
  feedbackId: string;
  onCancel?: () => void;
  onSubmitted?: () => void;
  parentId?: string;
}) => {
  const { data: session } = useSession();
  const createComment = useCreateTeamFeedbackComment(feedbackId);
  const editor = useEditor({
    content: "",
    editable: true,
    extensions: getCommentEditorExtensions({
      placeholder: parentId
        ? "Write a public reply..."
        : "Write a public comment...",
    }),
    immediatelyRender: false,
  });

  const handleSubmit = () => {
    if (!editor || editor.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }

    const body = serializeCommentToGitHubMarkdown(editor.getJSON());
    if (!body) return;

    editor.commands.clearContent();
    createComment.mutate(
      { body, parentId },
      {
        onError: () => {
          if (editor.isEmpty) editor.commands.setContent(body);
        },
        onSuccess: () => {
          onSubmitted?.();
        },
      },
    );
  };

  return (
    <Flex align="start" className={parentId ? "gap-2" : "mb-6 gap-2"}>
      <Box className="bg-surface flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar
          name={session?.user.name ?? undefined}
          size="xs"
          src={session?.user.image ?? undefined}
        />
      </Box>
      <Box className="min-w-0 flex-1">
        <Flex
          className={
            parentId
              ? "border-border/40 bg-surface-muted/40 min-h-16 rounded-xl border px-3 pb-3"
              : "border-border/40 bg-surface-muted/40 min-h-24 rounded-2xl border px-4 pb-4"
          }
          direction="column"
          gap={2}
          justify="between"
        >
          <TextEditor
            aria-label={parentId ? "Reply" : "Comment"}
            className="prose-base prose-a:text-foreground leading-6 antialiased"
            editor={editor}
          />
          <Flex gap={1} justify="end">
            {onCancel ? (
              <Button
                color="tertiary"
                onClick={onCancel}
                size="sm"
                variant="naked"
              >
                Cancel
              </Button>
            ) : null}
            <Button
              color="tertiary"
              onClick={handleSubmit}
              size="sm"
              variant="outline"
            >
              {parentId ? "Reply" : "Comment"}
            </Button>
          </Flex>
        </Flex>
      </Box>
    </Flex>
  );
};

const CommentRow = ({
  comment,
  feedbackId,
  isReply = false,
  replies = [],
}: {
  comment: TeamFeedbackComment;
  feedbackId: string;
  isReply?: boolean;
  replies?: TeamFeedbackComment[];
}) => {
  const [isReplying, setIsReplying] = useState(false);

  return (
    <Box
      className={
        isReply
          ? "border-border ml-9 border-l-2 pt-1 pb-3 pl-2"
          : "relative pb-5"
      }
    >
      <Flex align="center" gap={1}>
        <Box className="bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]">
          <Avatar
            className="relative top-0.5"
            name={comment.authorName}
            size="xs"
            src={comment.authorAvatar ?? undefined}
          />
        </Box>
        <Text className="ml-1 text-black dark:text-white">
          {comment.authorName}
        </Text>
        <Text className="mx-0.5 text-[0.95rem]" color="muted">
          ·
        </Text>
        <Text className="text-[0.95rem]" color="muted">
          <TimeAgo timestamp={comment.createdAt} />
        </Text>
      </Flex>
      <Box className="prose prose-stone dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-surface-muted prose-pre:text-foreground mt-0.5 ml-9 max-w-full leading-6">
        <Markdown remarkPlugins={[remarkGfm]}>{comment.body}</Markdown>
      </Box>
      {replies.length > 0 ? (
        <Box className="mt-2">
          {replies.map((reply) => (
            <CommentRow
              comment={reply}
              feedbackId={feedbackId}
              isReply
              key={reply.id}
            />
          ))}
        </Box>
      ) : null}
      {!isReply && !isReplying ? (
        <Button
          className="mt-2 ml-9 px-2"
          color="tertiary"
          leftIcon={<ReplyIcon className="h-4" />}
          onClick={() => {
            setIsReplying(true);
          }}
          size="sm"
          variant="naked"
        >
          Reply
        </Button>
      ) : null}
      {isReplying ? (
        <Box className="mt-3 ml-9">
          <FeedbackCommentComposer
            feedbackId={feedbackId}
            onCancel={() => {
              setIsReplying(false);
            }}
            onSubmitted={() => {
              setIsReplying(false);
            }}
            parentId={comment.id}
          />
        </Box>
      ) : null}
    </Box>
  );
};

export const FeedbackCommentThreads = ({
  comments,
  feedbackId,
}: {
  comments: TeamFeedbackComment[];
  feedbackId: string;
}) =>
  getCommentThreads(comments).map(({ comment, replies }) => (
    <CommentRow
      comment={comment}
      feedbackId={feedbackId}
      key={comment.id}
      replies={replies}
    />
  ));
