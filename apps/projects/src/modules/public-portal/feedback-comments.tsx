"use client";

import { useState, useTransition } from "react";
import { Avatar, Box, Button, Flex, Text, TextArea } from "ui";
import { CommentIcon } from "icons";
import { toast } from "sonner";
import type {
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestComment,
} from "./types";
import { createFeedbackCommentAction } from "./actions";
import { getPublicAvatarColor } from "./avatar-color";

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
  const [body, setBody] = useState("");
  const [isPending, startTransition] = useTransition();

  if (!viewer) {
    return (
      <Flex
        align="center"
        className="border-border/60 bg-surface-muted/40 rounded-xl border-[0.5px] px-4 py-3"
        justify="between"
      >
        <Text color="muted">Log in to join the conversation.</Text>
        <Button color="invert" href="/" size="sm">
          Login/signup
        </Button>
      </Flex>
    );
  }

  return (
    <Flex align="start" className="mb-6 gap-2">
      <Box className="bg-surface flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar name={viewer.name} size="xs" src={viewer.avatarUrl} />
      </Box>
      <Flex
        className="border-border/40 bg-surface-muted/40 min-h-24 min-w-0 flex-1 rounded-xl border px-4 pb-4"
        direction="column"
        gap={2}
        justify="between"
      >
        <TextArea
          aria-label="Comment"
          className="min-h-16 border-0 bg-transparent px-0 py-3 leading-6 focus-visible:ring-0 dark:bg-transparent"
          onChange={(event) => {
            setBody(event.target.value);
          }}
          placeholder="Leave a comment..."
          value={body}
        />
        <Flex justify="end">
          <Button
            color="tertiary"
            disabled={body.trim().length === 0 || isPending}
            onClick={() => {
              startTransition(async () => {
                const response = await createFeedbackCommentAction({
                  body,
                  itemId: request.id,
                  itemSlug: request.slug,
                  portalSlug: portal.slug,
                  workspaceSlug: portal.workspace.slug,
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
                setBody("");
              });
            }}
            size="sm"
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
      <Box className="bg-surface flex aspect-square items-center rounded-full p-[0.3rem]">
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
  const [comments, setComments] = useState(request.comments);
  const [visibleCount, setVisibleCount] = useState(COMMENTS_PAGE_SIZE);
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
        <Text as="span" color="muted">
          {comments.length}
        </Text>
      </Text>
      <FeedbackCommentComposer
        onCreated={(comment) => {
          setComments((current) => [...current, comment]);
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
