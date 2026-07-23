"use client";

import { useState } from "react";
import type { Editor } from "@tiptap/core";
import { useEditor } from "@tiptap/react";
import { Avatar, Box, Button, Flex, Text, TextEditor } from "ui";
import { CommentIcon, ReplyIcon } from "icons";
import { toast } from "sonner";
import { getStoryCommentEditorExtensions } from "@/modules/story/components/story-comment-editor";
import type {
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestComment,
} from "./types";
import { getPublicAvatarColor } from "./avatar-color";
import { getRequestLoginUrl } from "./utils";
import { useCreatePublicFeedbackComment } from "./feedback-mutations";

const COMMENTS_PAGE_SIZE = 10;

const FeedbackCommentComposer = ({
  onCancel,
  onSubmitted,
  parentId,
  portal,
  request,
  viewer,
}: {
  onCancel?: () => void;
  onSubmitted: () => void;
  parentId?: string;
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const editor = useEditor({
    content: "",
    editable: true,
    extensions: getStoryCommentEditorExtensions({
      placeholder: parentId ? "Reply to comment..." : "Leave a comment...",
    }),
    immediatelyRender: false,
  });

  if (!viewer) {
    return (
      <Flex
        align="center"
        className="border-border/60 bg-surface-muted/40 rounded-xl border-[0.5px] px-4 py-3"
        justify="between"
      >
        <Text color="muted">Log in to join the conversation.</Text>
        <Flex gap={1}>
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
            color="invert"
            href={getRequestLoginUrl(portal, request)}
            size="sm"
          >
            Login/signup
          </Button>
        </Flex>
      </Flex>
    );
  }

  return (
    <AuthenticatedFeedbackCommentComposer
      editor={editor}
      onCancel={onCancel}
      onSubmitted={onSubmitted}
      parentId={parentId}
      portal={portal}
      request={request}
      viewer={viewer}
    />
  );
};

const AuthenticatedFeedbackCommentComposer = ({
  editor,
  onCancel,
  onSubmitted,
  parentId,
  portal,
  request,
  viewer,
}: {
  editor: Editor | null;
  onCancel?: () => void;
  onSubmitted: () => void;
  parentId?: string;
  portal: PublicPortal;
  request: PublicRequest;
  viewer: PublicPortalViewer;
}) => {
  const createComment = useCreatePublicFeedbackComment({
    portalSlug: portal.slug,
    request,
    viewer,
  });

  return (
    <Flex align="start" className={parentId ? "gap-2" : "mb-6 gap-2"}>
      <Box className="bg-background flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar name={viewer.name} size="xs" src={viewer.avatarUrl} />
      </Box>
      <Flex
        className={
          parentId
            ? "border-border/40 bg-surface-muted/40 min-h-16 min-w-0 flex-1 rounded-xl border px-3 pb-3"
            : "border-border/40 bg-surface-muted/40 min-h-24 min-w-0 flex-1 rounded-2xl border px-4 pb-4"
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
            onClick={() => {
              if (!editor || editor.isEmpty) {
                toast.error("Comment is required", {
                  description: "Please enter a comment before submitting",
                });
                return;
              }
              const body = editor.getText();
              editor.commands.clearContent();
              createComment.mutate(
                { body, parentId },
                {
                  onError: () => {
                    if (editor.isEmpty) {
                      editor.commands.setContent(body);
                    }
                  },
                  onSuccess: onSubmitted,
                },
              );
            }}
            size="sm"
            variant="outline"
          >
            {parentId ? "Reply" : "Comment"}
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};

const getCommentThreads = (comments: PublicRequestComment[]) => {
  const repliesByParent = new Map<string, PublicRequestComment[]>();
  const topLevelComments: PublicRequestComment[] = [];

  for (const comment of comments) {
    if (comment.parentId) {
      const replies = repliesByParent.get(comment.parentId) ?? [];
      replies.push(comment);
      repliesByParent.set(comment.parentId, replies);
    } else {
      topLevelComments.push(comment);
    }
  }

  return topLevelComments.map((comment) => ({
    comment,
    replies: (repliesByParent.get(comment.id) ?? []).reverse(),
  }));
};

const FeedbackComment = ({
  comment,
  isReply = false,
  portal,
  replies = [],
  request,
  viewer,
}: {
  comment: PublicRequestComment;
  isReply?: boolean;
  portal: PublicPortal;
  replies?: PublicRequestComment[];
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const [isReplying, setIsReplying] = useState(false);

  return (
    <Box
      className={
        isReply ? "border-border ml-9 border-l-2 pt-1 pb-3 pl-2" : "pb-5"
      }
    >
      <Flex align="center" gap={1}>
        <Box className="bg-background flex aspect-square items-center rounded-full p-[0.3rem]">
          <Avatar
            name={comment.authorName}
            size="xs"
            src={comment.authorAvatar}
            style={{
              backgroundColor: getPublicAvatarColor(comment.authorName),
            }}
          />
        </Box>
        <Text className="ml-1">{comment.authorName}</Text>
        <Text className="mx-0.5 text-[0.95rem]" color="muted">
          ·
        </Text>
        <Text className="text-[0.95rem]" color="muted">
          {comment.createdAtLabel}
        </Text>
      </Flex>
      <Text className="mt-1 ml-9 leading-6" color="muted">
        {comment.body}
      </Text>
      {replies.length > 0 ? (
        <Box className="mt-2">
          {replies.map((reply) => (
            <FeedbackComment
              comment={reply}
              isReply
              key={reply.id}
              portal={portal}
              request={request}
              viewer={viewer}
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
            onCancel={() => {
              setIsReplying(false);
            }}
            onSubmitted={() => {
              setIsReplying(false);
            }}
            parentId={comment.id}
            portal={portal}
            request={request}
            viewer={viewer}
          />
        </Box>
      ) : null}
    </Box>
  );
};

export const FeedbackDiscussion = ({
  portal,
  request,
  viewer,
}: {
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const [visibleCount, setVisibleCount] = useState(COMMENTS_PAGE_SIZE);
  const commentThreads = getCommentThreads(request.comments);
  const visibleThreads = commentThreads.slice(0, visibleCount);
  const hasMore = visibleCount < commentThreads.length;

  return (
    <Box>
      <Text
        as="h2"
        className="mb-5 flex items-center gap-1.5"
        fontWeight="medium"
      >
        <CommentIcon className="h-[1.1rem]" />
        Comments
      </Text>
      <FeedbackCommentComposer
        onSubmitted={() => {
          setVisibleCount((current) => current + 1);
        }}
        portal={portal}
        request={request}
        viewer={viewer}
      />
      {visibleThreads.length > 0 ? (
        visibleThreads.map(({ comment, replies }) => (
          <FeedbackComment
            comment={comment}
            key={comment.id}
            portal={portal}
            replies={replies}
            request={request}
            viewer={viewer}
          />
        ))
      ) : (
        <Text className="py-5" color="muted">
          No comments yet. Start the conversation.
        </Text>
      )}
      {hasMore ? (
        <Button
          className="ml-6 px-3 text-[0.95rem]"
          color="tertiary"
          onClick={() => {
            setVisibleCount((current) => current + COMMENTS_PAGE_SIZE);
          }}
          size="sm"
          variant="naked"
        >
          Load more comments
        </Button>
      ) : null}
    </Box>
  );
};
