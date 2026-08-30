"use client";

import { useState } from "react";
import { ArrowLeft2Icon } from "icons";
import { Avatar, Box, Flex, Text } from "ui";
import type {
  PublicPortal,
  PublicRequest,
  PublicRequestComment,
} from "@/shared/feedback-widget/types";
import { createWidgetFeedbackCommentAction } from "../actions";
import type { WidgetSubmissionIdentity } from "./types";
import { DetailVoteControl, StatusBadge, WidgetBackButton } from "./widget-ui";

const Comment = ({ comment }: { comment: PublicRequestComment }) => (
  <Flex align="start" className="gap-3 py-3">
    <Avatar
      className="shrink-0"
      name={comment.authorName}
      size="sm"
      src={comment.authorAvatar}
    />
    <Box className="min-w-0 flex-1">
      <Flex align="center" className="gap-1.5">
        <Text className="text-[12px]" fontWeight="semibold">
          {comment.authorName}
        </Text>
        <Text className="text-[10px]" color="muted">
          · {comment.createdAtLabel}
        </Text>
      </Flex>
      <Text
        className="mt-1 text-[12px] leading-5 whitespace-pre-wrap"
        color="muted"
      >
        {comment.body}
      </Text>
    </Box>
  </Flex>
);

export const RequestDetail = ({
  canUseIdentity,
  identity,
  isWriteLocked,
  isVoting,
  onBack,
  onCommentCreated,
  onRequireIdentity,
  onVote,
  portal,
  request,
}: {
  canUseIdentity: (identity: WidgetSubmissionIdentity | null) => boolean;
  identity: WidgetSubmissionIdentity | null;
  isWriteLocked: boolean;
  isVoting: boolean;
  onBack: () => void;
  onCommentCreated: (comment: PublicRequestComment) => void;
  onRequireIdentity: () => void;
  onVote: (direction: -1 | 1) => void;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  let commentActionLabel = "Verify to comment";
  if (identity) {
    commentActionLabel = isSubmitting ? "Posting…" : "Comment";
  }

  const submitComment = async () => {
    if (!canUseIdentity(identity)) {
      setError("Wait for the new identity to finish before commenting.");
      return;
    }
    if (!identity) {
      onRequireIdentity();
      return;
    }
    if (!comment.trim() || isSubmitting) return;
    setIsSubmitting(true);
    setError("");
    const response = await createWidgetFeedbackCommentAction({
      body: comment,
      itemId: request.id,
      participantKind: identity.kind,
      portalSlug: portal.slug,
      sessionToken:
        identity.kind === "account" ? undefined : identity.sessionToken,
    })
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("Unable to add your comment");
      return;
    }
    if (!canUseIdentity(identity)) return;
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "Unable to add your comment");
      return;
    }
    setComment("");
    onCommentCreated(response.data);
  };

  return (
    <Box className="bg-background absolute inset-0 z-20 flex min-h-0 flex-col">
      <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
        <WidgetBackButton aria-label="Back to feedback" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetBackButton>
        <Text className="text-[16px]" fontWeight="semibold">
          Feedback
        </Text>
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-6 pb-8">
        <Flex align="center" className="mt-3 gap-3">
          <Avatar
            name={request.authorName || "Anonymous"}
            size="md"
            src={request.authorAvatar}
          />
          <Box>
            <Text className="text-[13px]" fontWeight="semibold">
              {request.authorName || "Anonymous"}
            </Text>
            <Text className="text-[11px]" color="muted">
              {request.createdAtLabel}
            </Text>
          </Box>
        </Flex>
        <Text
          as="h1"
          className="mt-6 text-[22px] leading-8"
          fontWeight="semibold"
        >
          {request.title}
        </Text>
        {request.description ? (
          <Text
            className="mt-3 text-[13px] leading-6 whitespace-pre-wrap"
            color="muted"
          >
            {request.description}
          </Text>
        ) : null}
        <Flex align="center" className="mt-6" justify="between">
          <StatusBadge status={request.status} />
          <DetailVoteControl
            disabled={isWriteLocked}
            isPending={isVoting}
            onVote={onVote}
            request={request}
          />
        </Flex>
        <Box className="border-border/70 mt-7 border-t pt-6">
          <Flex align="center" className="mb-3 gap-2">
            <Text
              className="text-[11px] tracking-[0.08em] uppercase"
              fontWeight="semibold"
            >
              Comments
            </Text>
            <Text className="text-[11px]" color="muted">
              {request.commentCount}
            </Text>
          </Flex>
          <Box className="border-border bg-surface rounded-xl border p-3">
            <textarea
              aria-label="Add a comment"
              className="text-foreground placeholder:text-text-muted/60 min-h-20 w-full resize-none border-0 bg-transparent p-0 text-[12px] leading-5 outline-none"
              maxLength={5000}
              onChange={(event) => {
                setComment(event.target.value);
              }}
              placeholder={
                identity ? "Add a comment…" : "Verify your email to comment…"
              }
              value={comment}
            />
            <Flex align="center" justify="end">
              <button
                className="bg-foreground text-background inline-flex h-9 items-center rounded-lg px-4 text-[12px] font-semibold disabled:cursor-not-allowed disabled:opacity-35"
                disabled={
                  isWriteLocked ||
                  (Boolean(identity) && (!comment.trim() || isSubmitting))
                }
                onClick={() => void submitComment()}
                type="button"
              >
                {commentActionLabel}
              </button>
            </Flex>
          </Box>
          {error ? (
            <Text className="mt-3 text-[11px] text-red-600 dark:text-red-400">
              {error}
            </Text>
          ) : null}
          <Box className="mt-3 divide-y divide-[var(--color-border)]/60">
            {request.comments.length > 0 ? (
              request.comments.map((item) => (
                <Comment comment={item} key={item.id} />
              ))
            ) : (
              <Text className="py-8 text-center text-[12px]" color="muted">
                No comments yet. Start the conversation.
              </Text>
            )}
          </Box>
        </Box>
      </Box>
    </Box>
  );
};
