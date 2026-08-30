"use client";

import { useState } from "react";
import { ArrowLeft2Icon, CheckIcon, ImageIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import type {
  PublicPortal,
  PublicRequest,
} from "@/shared/feedback-widget/types";
import {
  createWidgetFeedbackAction,
  type CreateWidgetFeedbackResult,
} from "../actions";
import type { WidgetSubmissionIdentity } from "./types";
import { WidgetBackButton, WidgetIconButton } from "./widget-ui";

export const FeedbackComposer = ({
  canUseIdentity,
  identity,
  isWriteLocked,
  onBack,
  onCreated,
  onRequireIdentity,
  portal,
}: {
  canUseIdentity: (identity: WidgetSubmissionIdentity | null) => boolean;
  identity: WidgetSubmissionIdentity | null;
  isWriteLocked: boolean;
  onBack: () => void;
  onCreated: (result: CreateWidgetFeedbackResult) => void;
  onRequireIdentity: () => void;
  portal: PublicPortal;
}) => {
  const [boardId, setBoardId] = useState(
    portal.boards.length === 1 ? portal.boards[0]?.id ?? "" : "",
  );
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showIdentityChoice, setShowIdentityChoice] = useState(false);

  const submit = async (activeIdentity: WidgetSubmissionIdentity | null) => {
    if (!boardId || !title.trim() || isSubmitting) return;
    if (!canUseIdentity(activeIdentity)) {
      setError(
        "Your feedback draft is safe. Wait for the new identity to finish before submitting.",
      );
      return;
    }
    setIsSubmitting(true);
    setError("");
    const participationIntent = activeIdentity?.kind ?? "anonymous";
    const response = await createWidgetFeedbackAction({
      boardId,
      description: description.trim(),
      participationIntent,
      portalSlug: portal.slug,
      sessionToken:
        activeIdentity && activeIdentity.kind !== "account"
          ? activeIdentity.sessionToken
          : undefined,
      title: title.trim(),
    })
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("Unable to submit feedback");
      return;
    }
    if (!canUseIdentity(activeIdentity)) return;
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "Unable to submit feedback");
      return;
    }
    if (response.data.participantKind !== participationIntent) {
      setError("The submission privacy setting could not be confirmed.");
      return;
    }
    onCreated(response.data);
  };

  if (showIdentityChoice && !identity) {
    return (
      <Box className="bg-background absolute inset-0 z-30 flex min-h-0 flex-col">
        <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
          <WidgetBackButton
            aria-label="Back to feedback draft"
            onClick={() => {
              setShowIdentityChoice(false);
            }}
          >
            <ArrowLeft2Icon className="h-5" />
          </WidgetBackButton>
          <Text className="text-[16px]" fontWeight="semibold">
            Choose how to submit
          </Text>
        </Flex>
        <Flex
          className="min-h-0 flex-1 px-6 py-8"
          direction="column"
          justify="between"
        >
          <Box>
            <Text className="text-[20px] leading-7" fontWeight="semibold">
              Get progress updates or stay anonymous
            </Text>
            <Text className="mt-2 text-[13px] leading-6" color="muted">
              Verify your email without creating an account to receive replies
              and meaningful status updates. Your feedback draft will stay here
              while you verify.
            </Text>
            <Button
              className="mt-6 w-full justify-center rounded-lg"
              color="invert"
              onClick={onRequireIdentity}
            >
              Continue with email
            </Button>
            {portal.participationMode === "anonymous_allowed" ? (
              <Button
                className="mt-3 w-full justify-center rounded-lg"
                color="tertiary"
                disabled={isSubmitting || isWriteLocked}
                onClick={() => {
                  void submit(null);
                }}
              >
                {isSubmitting ? "Posting…" : "Submit anonymously"}
              </Button>
            ) : null}
          </Box>
          <Text className="text-[11px] leading-5" color="muted">
            Anonymous submissions cannot receive personal replies or status
            emails. Administrators will not receive your name or email.
          </Text>
        </Flex>
      </Box>
    );
  }

  return (
    <Box className="bg-background absolute inset-0 z-30 flex min-h-0 flex-col">
      <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
        <WidgetBackButton aria-label="Back to feedback" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetBackButton>
        <Text className="text-[16px]" fontWeight="semibold">
          Share feedback
        </Text>
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-5 py-3">
        {portal.boards.length > 1 ? (
          <label className="mb-5 block space-y-2">
            <Text
              as="span"
              className="block text-[11px] tracking-[0.08em] uppercase"
              color="muted"
              fontWeight="semibold"
            >
              Board
            </Text>
            <select
              className="border-border bg-surface focus-visible:ring-ring h-10 w-full rounded-lg border px-3 text-[13px] outline-none focus-visible:ring-2"
              onChange={(event) => {
                setBoardId(event.target.value);
              }}
              value={boardId}
            >
              <option value="">Choose a board</option>
              {portal.boards.map((board) => (
                <option key={board.id} value={board.id}>
                  {board.name}
                </option>
              ))}
            </select>
          </label>
        ) : null}
        <input
          aria-label="Feedback title"
          className="text-foreground placeholder:text-text-muted/55 w-full border-0 bg-transparent p-0 text-[20px] leading-7 font-semibold outline-none"
          maxLength={200}
          onChange={(event) => {
            setTitle(event.target.value);
          }}
          placeholder="What could be better?"
          value={title}
        />
        <textarea
          aria-label="Feedback details"
          className="text-foreground placeholder:text-text-muted/50 mt-4 min-h-48 w-full resize-none border-0 bg-transparent p-0 text-[14px] leading-6 outline-none"
          maxLength={5000}
          onChange={(event) => {
            setDescription(event.target.value);
          }}
          placeholder="Tell us a little more about the problem, idea, or improvement."
          value={description}
        />
        {error ? (
          <Text className="mt-4 text-[12px] text-red-600 dark:text-red-400">
            {error}
          </Text>
        ) : null}
      </Box>
      <Flex align="center" className="shrink-0 px-5 py-4" justify="between">
        <WidgetIconButton aria-label="Add an image" disabled>
          <ImageIcon className="h-5" />
        </WidgetIconButton>
        <button
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-10 items-center justify-center rounded-lg px-5 text-[13px] font-semibold transition-opacity focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
          disabled={!boardId || !title.trim() || isSubmitting || isWriteLocked}
          onClick={() => {
            if (identity) {
              void submit(identity);
              return;
            }
            setShowIdentityChoice(true);
          }}
          type="button"
        >
          {isSubmitting ? "Submitting…" : "Submit feedback"}
        </button>
      </Flex>
    </Box>
  );
};

export const SubmissionSuccess = ({
  onView,
  request,
}: {
  onView: () => void;
  request: PublicRequest;
}) => (
  <Box className="bg-background absolute inset-0 z-40 flex flex-col">
    <Flex align="center" className="h-16 shrink-0 px-5">
      <Text className="text-[16px]" fontWeight="semibold">
        Feedback received
      </Text>
    </Flex>
    <Flex
      align="center"
      className="min-h-0 flex-1 px-10 text-center"
      direction="column"
      justify="center"
    >
      <Flex
        align="center"
        className="mb-5 size-12 rounded-full bg-emerald-500/12 text-emerald-600 dark:text-emerald-400"
        justify="center"
      >
        <CheckIcon className="h-5" />
      </Flex>
      <Text className="text-[19px] leading-7" fontWeight="semibold">
        Thanks for helping shape what comes next
      </Text>
      <Text className="mt-2 max-w-72 text-[12px] leading-5" color="muted">
        “{request.title}” is now in the feedback list.
      </Text>
      <Button
        className="mt-7 rounded-lg"
        color="invert"
        onClick={onView}
        size="sm"
      >
        View feedback
      </Button>
    </Flex>
  </Box>
);
