"use client";

import { useState, useTransition } from "react";
import { useEditor } from "@tiptap/react";
import { Avatar, Box, Button, Flex, Text, TextEditor } from "ui";
import { CommentIcon } from "icons";
import { toast } from "sonner";
import { getStoryCommentEditorExtensions } from "@/modules/story/components/story-comment-editor";
import type {
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestComment,
} from "./types";
import { createFeedbackCommentAction } from "./actions";
import { getPublicAvatarColor } from "./avatar-color";
import { getRequestLoginUrl } from "./utils";

const COMMENTS_PAGE_SIZE = 10;

const FeedbackCommentComposer = ({
  onCreated,
  portal,
  request,
  viewer,
}: {
  onCreated: (comment: PublicRequestComment) => void;
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const [isPending, startTransition] = useTransition();
  const editor = useEditor({
    content: "",
    editable: true,
    extensions: getStoryCommentEditorExtensions({
      placeholder: "Leave a comment...",
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
        <Button
          color="invert"
          href={getRequestLoginUrl(portal, request)}
          size="sm"
        >
          Login/signup
        </Button>
      </Flex>
    );
  }

  return (
    <Flex align="start" className="mb-6 gap-2">
      <Box className="bg-background flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar name={viewer.name} size="xs" src={viewer.avatarUrl} />
      </Box>
      <Flex
        className="border-border/40 bg-surface-muted/40 min-h-24 min-w-0 flex-1 rounded-2xl border px-4 pb-4"
        direction="column"
        gap={2}
        justify="between"
      >
        <TextEditor
          aria-label="Comment"
          className="prose-base prose-a:text-foreground leading-6 antialiased"
          editor={editor}
        />
        <Flex justify="end">
          <Button
            color="tertiary"
            disabled={isPending}
            onClick={() => {
              if (!editor || editor.isEmpty) {
                toast.error("Comment is required", {
                  description: "Please enter a comment before submitting",
                });
                return;
              }
              startTransition(async () => {
                const response = await createFeedbackCommentAction({
                  body: editor.getText(),
                  itemId: request.id,
                  itemSlug: request.slug,
                  portalSlug: portal.slug,
                });
                if (response.error?.message) {
                  toast.error("Comment", {
                    description: response.error.message,
                  });
                  return;
                }
                if (response.data) {
                  onCreated({
                    id: response.data.id,
                    authorName: response.data.authorName,
                    authorAvatar: response.data.authorAvatar,
                    body: response.data.body,
                    createdAtLabel: "Just now",
                  });
                }
                editor.commands.clearContent();
              });
            }}
            size="sm"
            variant="outline"
          >
            Comment
          </Button>
        </Flex>
      </Flex>
    </Flex>
  );
};

const FeedbackComment = ({ comment }: { comment: PublicRequestComment }) => (
  <Box className="pb-5">
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
  </Box>
);

export const FeedbackDiscussion = ({
  portal,
  request,
  viewer,
}: {
  portal: PublicPortal;
  request: PublicRequest;
  viewer?: PublicPortalViewer | null;
}) => {
  const [createdComments, setCreatedComments] = useState<
    PublicRequestComment[]
  >([]);
  const [visibleCount, setVisibleCount] = useState(COMMENTS_PAGE_SIZE);
  const comments = [...request.comments, ...createdComments].filter(
    (comment, index, allComments) =>
      allComments.findIndex(({ id }) => id === comment.id) === index,
  );
  const visibleComments = comments.slice(0, visibleCount);
  const hasMore = visibleCount < comments.length;

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
        onCreated={(comment) => {
          setCreatedComments((current) => [...current, comment]);
          setVisibleCount((current) => current + 1);
        }}
        portal={portal}
        request={request}
        viewer={viewer}
      />
      {visibleComments.length > 0 ? (
        visibleComments.map((comment) => (
          <FeedbackComment comment={comment} key={comment.id} />
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
