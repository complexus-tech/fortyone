"use client";

import { useId, useState } from "react";
import { useRouter } from "next/navigation";
import type { Editor } from "@tiptap/core";
import { useEditor } from "@tiptap/react";
import { Avatar, Box, Button, Checkbox, Flex, Text, TextEditor } from "ui";
import { CommentIcon, ReplyIcon } from "icons";
import { toast } from "sonner";
import { getStoryCommentEditorExtensions } from "@/modules/story/components/story-comment-editor";
import type {
  PublicPortal,
  PublicPortalParticipant,
  PublicRequest,
  PublicRequestComment,
} from "./types";
import { getRequestLoginUrl } from "./utils";
import { useCreatePublicFeedbackComment } from "./feedback-mutations";
import { FeedbackGuestVerificationDialog } from "./guest-verification";
import { canVerifyAsGuest, isContactableParticipant } from "./participant";
import { updateFeedbackFollowAction } from "./actions";

const COMMENTS_PAGE_SIZE = 10;

const FeedbackCommentComposer = ({
  onCancel,
  onParticipantChange,
  onSubmitted,
  parentId,
  participant,
  portal,
  request,
}: {
  onCancel?: () => void;
  onParticipantChange: (participant: PublicPortalParticipant) => void;
  onSubmitted: () => void;
  parentId?: string;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const [verificationOpen, setVerificationOpen] = useState(false);
  const editor = useEditor({
    content: "",
    editable: true,
    extensions: getStoryCommentEditorExtensions({
      placeholder: parentId ? "Reply to comment..." : "Leave a comment...",
    }),
    immediatelyRender: false,
  });

  if (!isContactableParticipant(participant)) {
    const guestParticipationEnabled = canVerifyAsGuest(
      portal.participationMode,
    );
    return (
      <>
        <Flex
          align="center"
          className="border-border/60 bg-surface-muted/40 rounded-xl border-[0.5px] px-4 py-3"
          justify="between"
        >
          <Text color="muted">
            {guestParticipationEnabled
              ? "Verify your email to join the conversation."
              : "Log in to join the conversation."}
          </Text>
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
              href={
                guestParticipationEnabled
                  ? undefined
                  : getRequestLoginUrl(portal, request)
              }
              onClick={
                guestParticipationEnabled
                  ? () => {
                      setVerificationOpen(true);
                    }
                  : undefined
              }
              size="sm"
            >
              {guestParticipationEnabled
                ? "Continue with email"
                : "Login/signup"}
            </Button>
          </Flex>
        </Flex>
        {guestParticipationEnabled ? (
          <FeedbackGuestVerificationDialog
            onOpenChange={setVerificationOpen}
            onVerified={onParticipantChange}
            open={verificationOpen}
            portal={portal}
            purpose="comment"
          />
        ) : null}
      </>
    );
  }

  return (
    <ContactableFeedbackCommentComposer
      editor={editor}
      onCancel={onCancel}
      onSubmitted={onSubmitted}
      parentId={parentId}
      participant={participant}
      portal={portal}
      request={request}
    />
  );
};

const ContactableFeedbackCommentComposer = ({
  editor,
  onCancel,
  onSubmitted,
  parentId,
  participant,
  portal,
  request,
}: {
  editor: Editor | null;
  onCancel?: () => void;
  onSubmitted: () => void;
  parentId?: string;
  participant: Exclude<PublicPortalParticipant, { kind: "anonymous" }>;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const [notifyOnUpdates, setNotifyOnUpdates] = useState(false);
  const notifyOnUpdatesId = useId();
  const createComment = useCreatePublicFeedbackComment({
    participant,
    portalSlug: portal.slug,
    request,
  });

  return (
    <Flex align="start" className={parentId ? "gap-2" : "mb-6 gap-2"}>
      <Box className="bg-background flex aspect-square shrink-0 items-center rounded-full p-[0.3rem]">
        <Avatar name={participant.name} size="xs" src={participant.avatarUrl} />
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
        <Flex align="center" className="flex-wrap gap-2" justify="between">
          <label
            className="text-text-muted flex cursor-pointer items-center gap-2 text-sm"
            htmlFor={notifyOnUpdatesId}
          >
            <Checkbox
              checked={notifyOnUpdates}
              disabled={createComment.isPending}
              id={notifyOnUpdatesId}
              onCheckedChange={(checked) => {
                setNotifyOnUpdates(checked === true);
              }}
            />
            Notify me about updates
          </label>
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
                    onSuccess: () => {
                      if (notifyOnUpdates) {
                        void updateFeedbackFollowAction({
                          following: true,
                          itemId: request.id,
                          itemSlug: request.slug,
                          participantKind: participant.kind,
                          portalSlug: portal.slug,
                        }).then((response) => {
                          if (response.error?.message) {
                            toast.error(
                              "Comment posted, but updates were not enabled",
                              {
                                description: response.error.message,
                              },
                            );
                          }
                        });
                      }
                      onSubmitted();
                    },
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
  onParticipantChange,
  participant,
  portal,
  replies = [],
  request,
}: {
  comment: PublicRequestComment;
  isReply?: boolean;
  onParticipantChange: (participant: PublicPortalParticipant) => void;
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  replies?: PublicRequestComment[];
  request: PublicRequest;
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
              onParticipantChange={onParticipantChange}
              participant={participant}
              portal={portal}
              request={request}
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
            onParticipantChange={onParticipantChange}
            onSubmitted={() => {
              setIsReplying(false);
            }}
            parentId={comment.id}
            participant={participant}
            portal={portal}
            request={request}
          />
        </Box>
      ) : null}
    </Box>
  );
};

export const FeedbackDiscussion = ({
  participant,
  portal,
  request,
}: {
  participant: PublicPortalParticipant;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const router = useRouter();
  const [participantOverride, setParticipantOverride] =
    useState<PublicPortalParticipant | null>(null);
  const activeParticipant = participantOverride ?? participant;
  const [visibleCount, setVisibleCount] = useState(COMMENTS_PAGE_SIZE);
  const commentThreads = getCommentThreads(request.comments);
  const visibleThreads = commentThreads.slice(0, visibleCount);
  const hasMore = visibleCount < commentThreads.length;
  const updateParticipant = (nextParticipant: PublicPortalParticipant) => {
    setParticipantOverride(nextParticipant);
    router.refresh();
  };

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
        onParticipantChange={updateParticipant}
        onSubmitted={() => {
          setVisibleCount((current) => current + 1);
        }}
        participant={activeParticipant}
        portal={portal}
        request={request}
      />
      {visibleThreads.length > 0 ? (
        visibleThreads.map(({ comment, replies }) => (
          <FeedbackComment
            comment={comment}
            key={comment.id}
            onParticipantChange={updateParticipant}
            participant={activeParticipant}
            portal={portal}
            replies={replies}
            request={request}
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
