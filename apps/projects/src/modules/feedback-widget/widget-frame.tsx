"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ArrowLeft2Icon,
  BellIcon,
  CheckIcon,
  CloseIcon,
  EditIcon,
  FeedbackIcon,
  ImageIcon,
  RoadmapIcon,
  SearchIcon,
  ThumbsUpIcon,
  UpdatesIcon,
} from "icons";
import { Avatar, Box, Button, Flex, Input, Switch, Text } from "ui";
import { cn, getReadableTextColor } from "lib";
import type {
  PublicPortal,
  PublicPortalTab,
  PublicPortalUpdate,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestComment,
  PublicRequestStatus,
} from "@/modules/public-portal/types";
import { getPublicAvatarColor } from "@/modules/public-portal/avatar-color";
import { requestStatusMeta } from "@/modules/public-portal/status";
import {
  confirmWidgetFeedbackVerificationAction,
  createWidgetFeedbackAction,
  createWidgetFeedbackCommentAction,
  exchangeWidgetIdentityAction,
  markWidgetFeedbackUpdatesSeenAction,
  requestWidgetFeedbackVerificationAction,
  revokeWidgetIdentityAction,
  toggleWidgetFeedbackVoteAction,
} from "./actions";
import type {
  CreateWidgetFeedbackResult,
  WidgetParticipantSession,
} from "./actions";
import {
  getTrustedWidgetOrigin,
  isFeedbackWidgetMessage,
  postFeedbackWidgetMessage,
  type FeedbackWidgetMode,
  type FeedbackWidgetTheme,
} from "./protocol";

type WidgetRoadmap = Record<
  "completed" | "in_progress" | "planned",
  PublicRequest[]
>;

type WidgetSubmissionIdentity =
  | { kind: "account" }
  | { kind: "external" | "verified_guest"; sessionToken: string };

type PendingIdentityAction =
  | { identityEpoch: number; type: "comment" }
  | { identityEpoch: number; requestId: string; type: "vote" }
  | { identityEpoch: number; type: "submit" };

const tabs = [
  { icon: FeedbackIcon, label: "Feedback", value: "feedback" },
  { icon: RoadmapIcon, label: "Roadmap", value: "roadmap" },
  { icon: UpdatesIcon, label: "Updates", value: "updates" },
] satisfies {
  icon: typeof FeedbackIcon;
  label: string;
  value: PublicPortalTab;
}[];

const roadmapSections = [
  { label: "In progress", status: "in_progress" },
  { label: "Planned", status: "planned" },
  { label: "Completed", status: "completed" },
] as const;

const replaceRequest = (
  requests: PublicRequest[],
  updatedRequest: PublicRequest,
) =>
  requests.map((request) =>
    request.id === updatedRequest.id ? updatedRequest : request,
  );

const statusAccent = (status: PublicRequestStatus) => {
  if (status === "completed") return "bg-emerald-500";
  if (status === "in_progress") return "bg-violet-500";
  if (status === "planned") return "bg-blue-500";
  if (status === "reviewing") return "bg-orange-500";
  if (status === "closed") return "bg-zinc-400";
  return "bg-amber-500";
};

const WidgetIconButton = ({
  children,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
  <button
    className={cn(
      "text-text-muted hover:bg-state-hover hover:text-foreground focus-visible:ring-ring inline-flex size-9 shrink-0 items-center justify-center rounded-full transition-colors focus-visible:ring-2 focus-visible:outline-none",
      className,
    )}
    type="button"
    {...props}
  >
    {children}
  </button>
);

const UnreadBadge = ({ count }: { count: number }) => {
  if (count <= 0) return null;

  return (
    <span
      aria-hidden="true"
      className="absolute -top-1 -right-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-red-600 px-1 text-[9px] leading-none font-semibold text-white tabular-nums"
    >
      {count > 99 ? "99+" : count}
    </span>
  );
};

const StatusBadge = ({ status }: { status: PublicRequestStatus }) => {
  const meta = requestStatusMeta[status];
  return (
    <span className="border-border/80 text-text-muted inline-flex h-7 items-center gap-2 rounded-full border px-2.5 text-[11px] font-medium">
      <span className={cn("size-2 rounded-full", statusAccent(status))} />
      {meta.label}
    </span>
  );
};

const VoteButton = ({
  disabled,
  isPending,
  onClick,
  request,
}: {
  disabled?: boolean;
  isPending?: boolean;
  onClick: () => void;
  request: PublicRequest;
}) => {
  const voted = request.viewerVote === 1;
  return (
    <button
      aria-label={voted ? "Remove upvote" : "Upvote feedback"}
      aria-pressed={voted}
      className={cn(
        "border-border/80 bg-background text-text-muted hover:border-foreground/25 hover:text-foreground focus-visible:ring-ring inline-flex h-8 min-w-14 shrink-0 items-center justify-center gap-1.5 rounded-full border px-2.5 text-[12px] font-semibold tabular-nums transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-60",
        { "border-foreground/20 text-foreground": voted },
      )}
      disabled={disabled || isPending}
      onClick={(event) => {
        event.stopPropagation();
        onClick();
      }}
      type="button"
    >
      <ThumbsUpIcon className="h-3.5" />
      {request.voteCount}
    </button>
  );
};

const EmptyState = ({
  body,
  icon: Icon,
  title,
}: {
  body: string;
  icon: typeof FeedbackIcon;
  title: string;
}) => (
  <Flex align="center" className="px-10 py-20 text-center" direction="column">
    <Flex
      align="center"
      className="bg-surface-muted text-text-muted mb-5 size-11 rounded-2xl"
      justify="center"
    >
      <Icon className="h-5" />
    </Flex>
    <Text className="text-[15px]" fontWeight="semibold">
      {title}
    </Text>
    <Text className="mt-1.5 max-w-64 text-[12px] leading-5" color="muted">
      {body}
    </Text>
  </Flex>
);

const FeedbackRow = ({
  isWriteLocked,
  isVoting,
  onOpen,
  onVote,
  portal,
  request,
}: {
  isWriteLocked: boolean;
  isVoting: boolean;
  onOpen: () => void;
  onVote: () => void;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = portal.boards.find(
    (candidate) => candidate.id === request.boardId,
  );
  return (
    <div className="border-border/60 hover:bg-state-hover/35 grid w-full grid-cols-[minmax(0,1fr)_auto] items-start gap-3 border-b px-5 py-4 transition-colors last:border-b-0">
      <button
        className="focus-visible:ring-ring grid min-w-0 grid-cols-[2.25rem_minmax(0,1fr)] gap-3 text-left focus-visible:ring-2 focus-visible:outline-none"
        onClick={onOpen}
        type="button"
      >
        <Avatar
          className="mt-0.5 shrink-0"
          name={request.authorName || "Anonymous"}
          size="sm"
          src={request.authorAvatar}
          style={{ backgroundColor: getPublicAvatarColor(request.authorName) }}
        />
        <Box className="min-w-0">
          <Text
            className="line-clamp-1 text-[14px] leading-5"
            fontWeight="semibold"
          >
            {request.title}
          </Text>
          {request.description ? (
            <Text
              className="mt-1 line-clamp-2 text-[12px] leading-5"
              color="muted"
            >
              {request.description}
            </Text>
          ) : null}
          <Flex align="center" className="mt-2 min-w-0 gap-1.5 text-[11px]">
            <Text className="truncate" fontWeight="medium">
              {request.authorName || "Anonymous"}
            </Text>
            <span className="text-text-muted">·</span>
            <Text color="muted">{request.createdAtLabel}</Text>
            {board ? (
              <>
                <span className="text-text-muted">·</span>
                <Text className="truncate" color="muted">
                  {board.name}
                </Text>
              </>
            ) : null}
          </Flex>
        </Box>
      </button>
      <VoteButton
        disabled={isWriteLocked}
        isPending={isVoting}
        onClick={onVote}
        request={request}
      />
    </div>
  );
};

const FeedbackComposer = ({
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
          <WidgetIconButton
            aria-label="Back to feedback draft"
            onClick={() => {
              setShowIdentityChoice(false);
            }}
          >
            <ArrowLeft2Icon className="h-5" />
          </WidgetIconButton>
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
              className="mt-6 w-full justify-center rounded-full"
              color="invert"
              onClick={onRequireIdentity}
            >
              Continue with email
            </Button>
            {portal.participationMode === "anonymous_allowed" ? (
              <Button
                className="mt-3 w-full justify-center rounded-full"
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
        <WidgetIconButton aria-label="Back to feedback" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetIconButton>
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
              className="border-border bg-surface focus-visible:ring-ring h-10 w-full rounded-xl border px-3 text-[13px] outline-none focus-visible:ring-2"
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
        {!identity ? (
          <Text className="mt-5 text-[11px] leading-5" color="muted">
            After writing, you can verify your email for updates or explicitly
            choose anonymous submission when the portal allows it.
          </Text>
        ) : null}
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
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-10 items-center justify-center rounded-full px-5 text-[13px] font-semibold transition-opacity focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
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

const SubmissionSuccess = ({
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
        className="mt-7 rounded-full"
        color="invert"
        onClick={onView}
        size="sm"
      >
        View feedback
      </Button>
    </Flex>
  </Box>
);

const IdentityGate = ({
  onBack,
  onVerified,
  portal,
}: {
  onBack: () => void;
  onVerified: (session: WidgetParticipantSession) => void;
  portal: PublicPortal;
}) => {
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [hideNamePublicly, setHideNamePublicly] = useState(
    portal.guestIdentityPolicy === "always_mask_guests",
  );
  const [step, setStep] = useState<"details" | "verify">("details");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  const requestCode = async () => {
    if (!email.trim() || isSubmitting) return;
    setIsSubmitting(true);
    setError("");
    const response = await requestWidgetFeedbackVerificationAction({
      displayName,
      email,
      hideNamePublicly,
      portalSlug: portal.slug,
    })
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("Unable to send a verification code");
      return;
    }
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "Unable to send a verification code");
      return;
    }
    setStep("verify");
  };

  const confirmCode = async () => {
    if (!code.trim() || isSubmitting) return;
    setIsSubmitting(true);
    setError("");
    const response = await confirmWidgetFeedbackVerificationAction({
      code,
      email,
      portalSlug: portal.slug,
    })
      .catch(() => null)
      .finally(() => {
        setIsSubmitting(false);
      });
    if (!response) {
      setError("That code could not be verified");
      return;
    }
    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "That code could not be verified");
      return;
    }
    onVerified(response.data);
  };

  const submitLabel =
    step === "details" ? "Send verification code" : "Verify and continue";
  const actionLabel = isSubmitting ? "Please wait…" : submitLabel;

  return (
    <Box className="bg-background absolute inset-0 z-40 flex min-h-0 flex-col">
      <Flex align="center" className="h-16 shrink-0 px-4" gap={2}>
        <WidgetIconButton aria-label="Go back" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetIconButton>
        <Text className="text-[16px]" fontWeight="semibold">
          {step === "details" ? "Join the conversation" : "Check your email"}
        </Text>
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-6 pt-5">
        {step === "details" ? (
          <>
            <Text className="text-[19px] leading-7" fontWeight="semibold">
              Verify once, then stay in the widget
            </Text>
            <Text className="mt-2 text-[12px] leading-5" color="muted">
              We’ll send a short code so you can vote, comment, and receive
              relevant feedback updates without creating a FortyOne account.
            </Text>
            <Box className="mt-7 space-y-4">
              <label className="block space-y-2" htmlFor="feedback-widget-name">
                <Text
                  as="span"
                  className="block text-[11px]"
                  color="muted"
                  fontWeight="semibold"
                >
                  Name
                </Text>
                <Input
                  id="feedback-widget-name"
                  onChange={(event) => {
                    setDisplayName(event.target.value);
                  }}
                  placeholder="Your name"
                  value={displayName}
                />
              </label>
              <label
                className="block space-y-2"
                htmlFor="feedback-widget-email"
              >
                <Text
                  as="span"
                  className="block text-[11px]"
                  color="muted"
                  fontWeight="semibold"
                >
                  Email
                </Text>
                <Input
                  autoFocus
                  id="feedback-widget-email"
                  onChange={(event) => {
                    setEmail(event.target.value);
                  }}
                  placeholder="you@company.com"
                  type="email"
                  value={email}
                />
              </label>
              {portal.guestIdentityPolicy === "allow_public_masking" ? (
                <Flex
                  align="center"
                  className="border-border/70 rounded-xl border p-3"
                  justify="between"
                >
                  <Box className="pr-4">
                    <Text className="text-[12px]" fontWeight="medium">
                      Post as Anonymous
                    </Text>
                    <Text
                      className="mt-0.5 text-[10px] leading-4"
                      color="muted"
                    >
                      Your email stays private either way.
                    </Text>
                  </Box>
                  <Switch
                    checked={hideNamePublicly}
                    onCheckedChange={setHideNamePublicly}
                  />
                </Flex>
              ) : null}
            </Box>
          </>
        ) : (
          <>
            <Text className="text-[19px] leading-7" fontWeight="semibold">
              Enter your verification code
            </Text>
            <Text className="mt-2 text-[12px] leading-5" color="muted">
              We sent a code to {email}. It may take a moment to arrive.
            </Text>
            <Input
              aria-label="Verification code"
              autoComplete="one-time-code"
              autoFocus
              className="mt-7 text-center tracking-[0.28em]"
              inputMode="numeric"
              maxLength={8}
              onChange={(event) => {
                setCode(event.target.value);
              }}
              placeholder="000000"
              value={code}
            />
            <button
              className="text-text-muted hover:text-foreground mt-4 text-[11px] transition-colors"
              onClick={() => {
                setCode("");
                setError("");
                setStep("details");
              }}
              type="button"
            >
              Use a different email
            </button>
          </>
        )}
        {error ? (
          <Text className="mt-4 text-[12px] text-red-600 dark:text-red-400">
            {error}
          </Text>
        ) : null}
      </Box>
      <Box className="shrink-0 px-6 py-5">
        <button
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-10 w-full items-center justify-center rounded-full px-5 text-[13px] font-semibold focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
          disabled={
            step === "details"
              ? !email.trim() || isSubmitting
              : !code.trim() || isSubmitting
          }
          onClick={() =>
            void (step === "details" ? requestCode() : confirmCode())
          }
          type="button"
        >
          {actionLabel}
        </button>
      </Box>
    </Box>
  );
};

const Comment = ({ comment }: { comment: PublicRequestComment }) => (
  <Flex align="start" className="gap-3 py-3">
    <Avatar
      className="shrink-0"
      name={comment.authorName}
      size="sm"
      src={comment.authorAvatar}
      style={{ backgroundColor: getPublicAvatarColor(comment.authorName) }}
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

const RequestDetail = ({
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
  onVote: () => void;
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
        <WidgetIconButton aria-label="Back to feedback" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetIconButton>
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
            style={{
              backgroundColor: getPublicAvatarColor(request.authorName),
            }}
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
          <VoteButton
            disabled={isWriteLocked}
            isPending={isVoting}
            onClick={onVote}
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
          <Box className="border-border bg-surface rounded-2xl border p-3">
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
                className="bg-foreground text-background inline-flex h-8 items-center rounded-full px-4 text-[11px] font-semibold disabled:cursor-not-allowed disabled:opacity-35"
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

const UpdateDetail = ({
  onBack,
  update,
}: {
  onBack: () => void;
  update: PublicPortalUpdate;
}) => (
  <Box className="bg-background absolute inset-0 z-20 flex min-h-0 flex-col">
    <Flex align="center" className="h-16 shrink-0 px-4">
      <WidgetIconButton aria-label="Back to updates" onClick={onBack}>
        <ArrowLeft2Icon className="h-5" />
      </WidgetIconButton>
    </Flex>
    <Box className="min-h-0 flex-1 overflow-y-auto px-6 pb-10">
      <Text
        className="text-[11px] tracking-[0.09em] uppercase"
        color="muted"
        fontWeight="semibold"
      >
        {update.publishedAtLabel} · Update
      </Text>
      <Text
        as="h1"
        className="mt-3 text-[23px] leading-8"
        fontWeight="semibold"
      >
        {update.title}
      </Text>
      {update.summary ? (
        <Text className="mt-4 text-[14px] leading-6" color="muted">
          {update.summary}
        </Text>
      ) : null}
      <Box className="border-border/70 mt-7 border-t pt-7">
        <Text
          className="text-[13px] leading-6 whitespace-pre-wrap"
          color="muted"
        >
          {update.body}
        </Text>
      </Box>
      {update.linkedItems.length > 0 ? (
        <Box className="border-border bg-surface mt-7 rounded-2xl border p-4">
          <Text
            className="text-[11px] tracking-[0.08em] uppercase"
            color="muted"
            fontWeight="semibold"
          >
            Related feedback
          </Text>
          {update.linkedItems.map((item) => (
            <Flex align="center" className="mt-3 gap-3" key={item.id}>
              <span
                className={cn("size-2 rounded-full", statusAccent(item.status))}
              />
              <Text className="min-w-0 flex-1 text-[12px]" fontWeight="medium">
                {item.title}
              </Text>
            </Flex>
          ))}
        </Box>
      ) : null}
    </Box>
  </Box>
);

const UpdatesList = ({
  onOpen,
  updates,
}: {
  onOpen: (update: PublicPortalUpdate) => void;
  updates: PublicPortalUpdate[];
}) => {
  if (updates.length === 0) {
    return (
      <EmptyState
        body="Product news and shipped improvements will appear here."
        icon={UpdatesIcon}
        title="No updates yet"
      />
    );
  }

  return (
    <Box className="px-5 py-3">
      {updates.map((update) => (
        <button
          className="border-border/70 hover:bg-state-hover/35 focus-visible:ring-ring w-full border-b py-5 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
          key={update.id}
          onClick={() => {
            onOpen(update);
          }}
          type="button"
        >
          <Text
            className="text-[10px] tracking-[0.09em] uppercase"
            color="muted"
            fontWeight="semibold"
          >
            {update.publishedAtLabel} · Update
          </Text>
          <Text className="mt-2 text-[16px] leading-6" fontWeight="semibold">
            {update.title}
          </Text>
          <Text
            className="mt-2 line-clamp-3 text-[12px] leading-5"
            color="muted"
          >
            {update.summary || update.body}
          </Text>
        </button>
      ))}
    </Box>
  );
};

const BottomNavigation = ({
  activeTab,
  onSelect,
  showUpdates,
  unreadUpdateCount,
}: {
  activeTab: PublicPortalTab;
  onSelect: (tab: PublicPortalTab) => void;
  showUpdates: boolean;
  unreadUpdateCount: number;
}) => (
  <nav
    aria-label="Feedback sections"
    className={cn(
      "border-border/70 bg-background/96 grid shrink-0 border-t px-3 py-2 backdrop-blur-xl",
      showUpdates ? "grid-cols-3" : "grid-cols-2",
    )}
  >
    {tabs
      .filter((tab) => tab.value !== "updates" || showUpdates)
      .map((tab) => {
        const Icon = tab.icon;
        const active = activeTab === tab.value;
        const unreadCount = tab.value === "updates" ? unreadUpdateCount : 0;
        return (
          <button
            aria-current={active ? "page" : undefined}
            aria-label={
              unreadCount > 0
                ? `${tab.label}, ${unreadCount} unread ${unreadCount === 1 ? "update" : "updates"}`
                : tab.label
            }
            className={cn(
              "text-text-muted hover:text-foreground focus-visible:ring-ring flex h-11 flex-col items-center justify-center gap-1 rounded-xl text-[10px] font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none",
              { "text-foreground": active },
            )}
            key={tab.value}
            onClick={() => {
              onSelect(tab.value);
            }}
            type="button"
          >
            <span className="relative">
              <Icon className="h-[18px]" />
              <UnreadBadge count={unreadCount} />
            </span>
            {tab.label}
          </button>
        );
      })}
  </nav>
);

export const FeedbackWidgetFrame = ({
  initialTab,
  instanceId,
  mode,
  parentOrigin,
  portal,
  roadmap,
  theme,
  viewer,
}: {
  initialTab: PublicPortalTab;
  instanceId: string;
  mode: FeedbackWidgetMode;
  parentOrigin: string;
  portal: PublicPortal;
  roadmap: WidgetRoadmap;
  theme: FeedbackWidgetTheme;
  viewer?: PublicPortalViewer | null;
}) => {
  const rootRef = useRef<HTMLDivElement | null>(null);
  const identityExchangeRef = useRef(0);
  const [identity, setIdentity] = useState<WidgetSubmissionIdentity | null>(
    () => (viewer ? { kind: "account" } : null),
  );
  const activeIdentityRef = useRef(identity);
  const identityPendingRef = useRef(false);
  const updatesSeenTokenRef = useRef<string | null>(null);
  const [activeTab, setActiveTab] = useState(
    initialTab === "updates" && !portal.hasPublishedUpdates
      ? "feedback"
      : initialTab,
  );
  const [requests, setRequests] = useState(portal.requests);
  const [search, setSearch] = useState("");
  const [selectedRequest, setSelectedRequest] = useState<PublicRequest | null>(
    null,
  );
  const [selectedUpdate, setSelectedUpdate] =
    useState<PublicPortalUpdate | null>(null);
  const [isComposing, setIsComposing] = useState(false);
  const [composerIdentity, setComposerIdentity] =
    useState<WidgetSubmissionIdentity | null>(null);
  const [submissionSuccess, setSubmissionSuccess] =
    useState<CreateWidgetFeedbackResult | null>(null);
  const [identityError, setIdentityError] = useState("");
  const [isIdentityPending, setIsIdentityPending] = useState(false);
  const [unreadUpdateCount, setUnreadUpdateCount] = useState(0);
  const [pendingIdentityAction, setPendingIdentityAction] =
    useState<PendingIdentityAction | null>(null);
  const [votingRequestId, setVotingRequestId] = useState<string | null>(null);
  const [isDark, setIsDark] = useState(theme === "dark");
  const trustedParentOrigin = getTrustedWidgetOrigin(parentOrigin);

  const revokeContributorIdentity = useCallback(
    (identityToRevoke: WidgetSubmissionIdentity | null) => {
      if (!identityToRevoke || identityToRevoke.kind === "account") return;
      void revokeWidgetIdentityAction({
        portalSlug: portal.slug,
        sessionToken: identityToRevoke.sessionToken,
      }).catch(() => undefined);
    },
    [portal.slug],
  );

  const beginIdentityTransition = useCallback(
    (pending: boolean) => {
      const previousIdentity = activeIdentityRef.current;
      const exchangeId = identityExchangeRef.current + 1;
      identityExchangeRef.current = exchangeId;
      identityPendingRef.current = pending;
      activeIdentityRef.current = null;
      updatesSeenTokenRef.current = null;
      setIdentity(null);
      setComposerIdentity(null);
      setIdentityError("");
      setIsIdentityPending(pending);
      setPendingIdentityAction(null);
      setUnreadUpdateCount(0);
      setVotingRequestId(null);
      revokeContributorIdentity(previousIdentity);
      return exchangeId;
    },
    [revokeContributorIdentity],
  );

  const activateIdentity = useCallback(
    (nextIdentity: WidgetSubmissionIdentity, nextUnreadUpdateCount = 0) => {
      identityPendingRef.current = false;
      activeIdentityRef.current = nextIdentity;
      updatesSeenTokenRef.current = null;
      setIdentity(nextIdentity);
      setComposerIdentity(nextIdentity);
      setIsIdentityPending(false);
      setUnreadUpdateCount(Math.max(0, nextUnreadUpdateCount));
    },
    [],
  );

  const canUseIdentity = useCallback(
    (candidate: WidgetSubmissionIdentity | null) =>
      !identityPendingRef.current && activeIdentityRef.current === candidate,
    [],
  );

  const filteredRequests = useMemo(() => {
    const value = search.trim().toLowerCase();
    if (!value) return requests;
    return requests.filter((request) =>
      `${request.title} ${request.description} ${request.authorName}`
        .toLowerCase()
        .includes(value),
    );
  }, [requests, search]);

  const syncRequest = useCallback((updatedRequest: PublicRequest) => {
    setRequests((current) => replaceRequest(current, updatedRequest));
    setSelectedRequest((current) =>
      current?.id === updatedRequest.id ? updatedRequest : current,
    );
  }, []);

  const vote = useCallback(
    async (
      request: PublicRequest,
      activeIdentity: WidgetSubmissionIdentity,
    ) => {
      if (votingRequestId === request.id || !canUseIdentity(activeIdentity))
        return;
      const identityEpoch = identityExchangeRef.current;
      setVotingRequestId(request.id);
      const response = await toggleWidgetFeedbackVoteAction({
        itemId: request.id,
        participantKind: activeIdentity.kind,
        portalSlug: portal.slug,
        sessionToken:
          activeIdentity.kind === "account"
            ? undefined
            : activeIdentity.sessionToken,
        vote: 1,
      }).catch(() => null);
      if (
        identityExchangeRef.current !== identityEpoch ||
        !canUseIdentity(activeIdentity)
      )
        return;
      setVotingRequestId(null);
      if (!response) return;
      if (!response.data) return;
      syncRequest({
        ...request,
        viewerVote: response.data.vote,
        voteCount: response.data.voteCount,
      });
    },
    [canUseIdentity, portal.slug, syncRequest, votingRequestId],
  );

  const requestVote = useCallback(
    (request: PublicRequest) => {
      if (identityPendingRef.current) return;
      if (!identity) {
        setPendingIdentityAction({
          identityEpoch: identityExchangeRef.current,
          requestId: request.id,
          type: "vote",
        });
        return;
      }
      void vote(request, identity);
    },
    [identity, vote],
  );

  useEffect(() => {
    if (theme !== "auto") {
      setIsDark(theme === "dark");
      return;
    }
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const update = () => {
      setIsDark(media.matches);
    };
    update();
    media.addEventListener("change", update);
    return () => {
      media.removeEventListener("change", update);
    };
  }, [theme]);

  useEffect(() => {
    if (!trustedParentOrigin) return;
    postFeedbackWidgetMessage("ready", instanceId, trustedParentOrigin);
    const handleMessage = (event: MessageEvent) => {
      if (
        event.origin !== trustedParentOrigin ||
        event.source !== window.parent ||
        !isFeedbackWidgetMessage(event.data, instanceId)
      )
        return;

      if (event.data.event === "host-close") {
        setIsComposing(false);
        setComposerIdentity(null);
        setPendingIdentityAction(null);
        setSelectedRequest(null);
        setSelectedUpdate(null);
      }
      if (event.data.event === "host-identity-clear") {
        const requestIdValue = event.data.payload?.requestId;
        const requestId =
          typeof requestIdValue === "string" ||
          typeof requestIdValue === "number"
            ? String(requestIdValue)
            : undefined;
        beginIdentityTransition(false);
        postFeedbackWidgetMessage(
          "identity-cleared",
          instanceId,
          trustedParentOrigin,
          requestId ? { requestId } : undefined,
        );
      }
      if (event.data.event === "host-identify") {
        const requestIdValue = event.data.payload?.requestId;
        const requestId =
          typeof requestIdValue === "string" ||
          typeof requestIdValue === "number"
            ? String(requestIdValue)
            : undefined;
        const exchangeId = beginIdentityTransition(true);
        const assertion = event.data.payload?.assertion;
        if (typeof assertion !== "string" || assertion.length > 16384) {
          identityPendingRef.current = false;
          setIsIdentityPending(false);
          setIdentityError(
            "The product supplied an invalid customer identity.",
          );
          postFeedbackWidgetMessage(
            "identity-error",
            instanceId,
            trustedParentOrigin,
            {
              message: "invalid_assertion",
              ...(requestId ? { requestId } : {}),
            },
          );
          return;
        }
        void exchangeWidgetIdentityAction({
          assertion,
          parentOrigin: trustedParentOrigin,
          portalSlug: portal.slug,
        }).then(
          (response) => {
            if (identityExchangeRef.current !== exchangeId) {
              if (response.data) {
                revokeContributorIdentity({
                  kind: response.data.participant.kind,
                  sessionToken: response.data.session.token,
                });
              }
              return;
            }
            identityPendingRef.current = false;
            setIsIdentityPending(false);
            if (!response.data) {
              setIdentityError(
                response.error?.message ??
                  "Your customer identity could not be verified. Choose email verification or anonymous submission explicitly.",
              );
              postFeedbackWidgetMessage(
                "identity-error",
                instanceId,
                trustedParentOrigin,
                {
                  message: "exchange_failed",
                  ...(requestId ? { requestId } : {}),
                },
              );
              return;
            }
            activateIdentity(
              {
                kind: response.data.participant.kind,
                sessionToken: response.data.session.token,
              },
              response.data.participant.unreadUpdateCount,
            );
            postFeedbackWidgetMessage(
              "identity-ready",
              instanceId,
              trustedParentOrigin,
              {
                kind: response.data.participant.kind,
                ...(requestId ? { requestId } : {}),
              },
            );
          },
          () => {
            if (identityExchangeRef.current !== exchangeId) return;
            identityPendingRef.current = false;
            setIsIdentityPending(false);
            setIdentityError(
              "Your customer identity could not be verified. Choose email verification or anonymous submission explicitly.",
            );
            postFeedbackWidgetMessage(
              "identity-error",
              instanceId,
              trustedParentOrigin,
              {
                message: "exchange_failed",
                ...(requestId ? { requestId } : {}),
              },
            );
          },
        );
      }
    };
    window.addEventListener("message", handleMessage);
    return () => {
      window.removeEventListener("message", handleMessage);
    };
  }, [
    activateIdentity,
    beginIdentityTransition,
    instanceId,
    portal.slug,
    revokeContributorIdentity,
    trustedParentOrigin,
  ]);

  useEffect(() => {
    if (
      activeTab !== "updates" ||
      unreadUpdateCount <= 0 ||
      !identity ||
      identity.kind === "account" ||
      identityPendingRef.current ||
      updatesSeenTokenRef.current === identity.sessionToken
    )
      return;

    const identityEpoch = identityExchangeRef.current;
    const sessionToken = identity.sessionToken;
    updatesSeenTokenRef.current = sessionToken;
    void markWidgetFeedbackUpdatesSeenAction({
      portalSlug: portal.slug,
      sessionToken,
    }).then(
      (response) => {
        if (
          identityExchangeRef.current !== identityEpoch ||
          activeIdentityRef.current !== identity
        )
          return;
        if (!response.error) {
          setUnreadUpdateCount(0);
          return;
        }
        updatesSeenTokenRef.current = null;
      },
      () => {
        if (
          identityExchangeRef.current === identityEpoch &&
          activeIdentityRef.current === identity
        ) {
          updatesSeenTokenRef.current = null;
        }
      },
    );
  }, [activeTab, identity, portal.slug, unreadUpdateCount]);

  useEffect(() => {
    if (!trustedParentOrigin || mode === "inline") return;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        postFeedbackWidgetMessage("escape", instanceId, trustedParentOrigin);
      }
    };
    window.addEventListener("keydown", handleEscape);
    return () => {
      window.removeEventListener("keydown", handleEscape);
    };
  }, [instanceId, mode, trustedParentOrigin]);

  useEffect(() => {
    if (!trustedParentOrigin || mode !== "inline" || !rootRef.current) return;
    const root = rootRef.current;
    const observer = new ResizeObserver(() => {
      postFeedbackWidgetMessage("resize", instanceId, trustedParentOrigin, {
        height: Math.ceil(root.scrollHeight),
      });
    });
    observer.observe(root);
    return () => {
      observer.disconnect();
    };
  }, [instanceId, mode, trustedParentOrigin]);

  const requestClose = () => {
    if (trustedParentOrigin) {
      postFeedbackWidgetMessage("close", instanceId, trustedParentOrigin);
    }
  };

  const openExternal = (path: string) => {
    if (!trustedParentOrigin) return;
    postFeedbackWidgetMessage(
      "open-external",
      instanceId,
      trustedParentOrigin,
      { href: new URL(path, window.location.origin).href },
    );
  };

  const openComposer = () => {
    if (identityPendingRef.current) return;
    const currentIdentity = activeIdentityRef.current;
    if (currentIdentity || portal.participationMode !== "account_required") {
      setComposerIdentity(currentIdentity);
      setIsComposing(true);
      return;
    }
    openExternal(`/portal/${portal.slug}/feedback?newFeedback=true`);
  };

  const handleVerified = (session: WidgetParticipantSession) => {
    const pendingAction = pendingIdentityAction;
    if (
      !pendingAction ||
      identityExchangeRef.current !== pendingAction.identityEpoch ||
      identityPendingRef.current
    ) {
      revokeContributorIdentity({
        kind: session.participant.kind,
        sessionToken: session.session.token,
      });
      return;
    }
    identityExchangeRef.current += 1;
    const verifiedIdentity: WidgetSubmissionIdentity = {
      kind: session.participant.kind,
      sessionToken: session.session.token,
    };
    activateIdentity(verifiedIdentity, session.participant.unreadUpdateCount);
    setPendingIdentityAction(null);
    if (pendingAction.type === "submit") {
      setComposerIdentity(verifiedIdentity);
      setIsComposing(true);
      return;
    }
    if (pendingAction.type === "vote") {
      const request = requests.find(
        (item) => item.id === pendingAction.requestId,
      );
      if (request) void vote(request, verifiedIdentity);
    }
  };
  return (
    <div
      className={cn(
        "bg-background text-foreground relative flex h-dvh min-h-0 w-full flex-col overflow-hidden antialiased",
        { dark: isDark },
      )}
      ref={rootRef}
    >
      <Flex
        align="center"
        className="bg-background/96 relative z-10 h-16 shrink-0 gap-3 px-5 backdrop-blur-xl"
      >
        <Avatar
          className="!size-8 text-[12px] font-bold"
          name={portal.workspace.name}
          rounded="lg"
          size="sm"
          src={portal.workspace.avatarUrl}
          style={{
            backgroundColor: portal.workspace.color,
            color: getReadableTextColor(portal.workspace.color),
          }}
        />
        <Text
          className="min-w-0 flex-1 truncate text-[15px]"
          fontWeight="semibold"
        >
          {portal.workspace.name}
        </Text>
        <button
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-9 shrink-0 items-center gap-2 rounded-full px-4 text-[12px] font-semibold shadow-sm focus-visible:ring-2 focus-visible:outline-none"
          disabled={isIdentityPending}
          onClick={openComposer}
          type="button"
        >
          <EditIcon className="h-3.5" />
          Add feedback
        </button>
        {identity && portal.hasPublishedUpdates ? (
          <WidgetIconButton
            aria-label={
              unreadUpdateCount > 0
                ? `View feedback updates, ${unreadUpdateCount} unread ${unreadUpdateCount === 1 ? "update" : "updates"}`
                : "View feedback updates"
            }
            className="relative hidden min-[380px]:inline-flex"
            onClick={() => {
              setActiveTab("updates");
              setSelectedRequest(null);
              setSelectedUpdate(null);
            }}
          >
            <BellIcon className="h-[18px]" />
            <UnreadBadge count={unreadUpdateCount} />
          </WidgetIconButton>
        ) : null}
        {mode !== "inline" ? (
          <WidgetIconButton
            aria-label="Close feedback widget"
            onClick={requestClose}
          >
            <CloseIcon className="h-5" />
          </WidgetIconButton>
        ) : null}
      </Flex>
      {isIdentityPending ? (
        <Text
          className="border-border/70 border-t px-5 py-2 text-[11px]"
          color="muted"
        >
          Verifying your customer identity…
        </Text>
      ) : null}
      {identityError ? (
        <Text className="border-border/70 border-t px-5 py-2 text-[11px] text-red-600 dark:text-red-400">
          {identityError}
        </Text>
      ) : null}

      <Box className="relative min-h-0 flex-1 overflow-y-auto">
        {activeTab === "feedback" ? (
          <Box>
            <Box className="bg-background sticky top-0 z-10 px-5 pt-4 pb-3">
              <div className="border-border bg-surface relative flex h-11 items-center rounded-full border px-4">
                <SearchIcon className="text-text-muted h-4 shrink-0" />
                <input
                  aria-label="Search feedback"
                  className="text-foreground placeholder:text-text-muted/65 h-full min-w-0 flex-1 border-0 bg-transparent px-3 text-[12px] outline-none"
                  onChange={(event) => {
                    setSearch(event.target.value);
                  }}
                  placeholder="Search feedback"
                  type="search"
                  value={search}
                />
              </div>
            </Box>
            {filteredRequests.length > 0 ? (
              filteredRequests.map((request) => (
                <FeedbackRow
                  isVoting={votingRequestId === request.id}
                  isWriteLocked={isIdentityPending}
                  key={request.id}
                  onOpen={() => {
                    setSelectedRequest(request);
                  }}
                  onVote={() => {
                    requestVote(request);
                  }}
                  portal={portal}
                  request={request}
                />
              ))
            ) : (
              <EmptyState
                body={
                  search
                    ? "Try a different phrase or share the idea yourself."
                    : "New suggestions will appear here as soon as they are shared."
                }
                icon={FeedbackIcon}
                title={search ? "No matching feedback" : "No feedback yet"}
              />
            )}
          </Box>
        ) : null}

        {activeTab === "roadmap" ? (
          <Box className="space-y-9 px-5 py-6">
            {roadmapSections.map((section) => {
              const items = roadmap[section.status];
              return (
                <Box key={section.status}>
                  <Flex align="center" className="mb-3" justify="between">
                    <Flex align="center" className="gap-2.5">
                      <span
                        className={cn(
                          "size-3 rounded-full",
                          statusAccent(section.status),
                        )}
                      />
                      <Text className="text-[15px]" fontWeight="semibold">
                        {section.label}
                      </Text>
                    </Flex>
                    <Text className="text-[11px] tabular-nums" color="muted">
                      {String(items.length).padStart(2, "0")}
                    </Text>
                  </Flex>
                  {items.length > 0 ? (
                    <Box className="border-border/70 ml-1.5 border-l">
                      {items.map((request) => (
                        <div
                          className="hover:bg-state-hover/35 relative grid w-full grid-cols-[minmax(0,1fr)_auto] items-start gap-3 py-3 pr-1 pl-6 transition-colors"
                          key={request.id}
                        >
                          <span
                            className={cn(
                              "ring-background absolute top-5 -left-[5px] size-2.5 rounded-full ring-4",
                              statusAccent(section.status),
                            )}
                          />
                          <button
                            className="focus-visible:ring-ring min-w-0 text-left focus-visible:ring-2 focus-visible:outline-none"
                            onClick={() => {
                              setSelectedRequest(request);
                            }}
                            type="button"
                          >
                            <Box className="min-w-0">
                              <Text
                                className="line-clamp-1 text-[13px]"
                                fontWeight="semibold"
                              >
                                {request.title}
                              </Text>
                              {request.description ? (
                                <Text
                                  className="mt-1 line-clamp-2 text-[11px] leading-5"
                                  color="muted"
                                >
                                  {request.description}
                                </Text>
                              ) : null}
                              <Text className="mt-2 text-[11px]" color="muted">
                                {request.authorName}
                              </Text>
                            </Box>
                          </button>
                          <VoteButton
                            disabled={isIdentityPending}
                            isPending={votingRequestId === request.id}
                            onClick={() => {
                              requestVote(request);
                            }}
                            request={request}
                          />
                        </div>
                      ))}
                    </Box>
                  ) : (
                    <Text className="ml-5 py-3 text-[11px]" color="muted">
                      Nothing here yet
                    </Text>
                  )}
                </Box>
              );
            })}
          </Box>
        ) : null}

        {activeTab === "updates" ? (
          <UpdatesList onOpen={setSelectedUpdate} updates={portal.updates} />
        ) : null}
      </Box>

      <BottomNavigation
        activeTab={activeTab}
        onSelect={(tab) => {
          setActiveTab(tab);
          setSelectedRequest(null);
          setSelectedUpdate(null);
        }}
        showUpdates={portal.hasPublishedUpdates}
        unreadUpdateCount={unreadUpdateCount}
      />
      <button
        className="border-border/70 text-text-muted hover:text-foreground bg-background focus-visible:ring-ring flex h-11 shrink-0 items-center justify-center border-t px-5 py-3 text-[10px] transition-colors focus-visible:ring-2 focus-visible:outline-none"
        onClick={() => {
          if (!trustedParentOrigin) return;
          postFeedbackWidgetMessage(
            "open-external",
            instanceId,
            trustedParentOrigin,
            {
              href: new URL(`/portal/${portal.slug}`, window.location.origin)
                .href,
            },
          );
        }}
        type="button"
      >
        Powered by{" "}
        <span className="ml-1 font-semibold text-[var(--color-foreground)]">
          FortyOne
        </span>
        <span aria-hidden="true" className="ml-1">
          ↗
        </span>
      </button>

      {selectedRequest ? (
        <RequestDetail
          canUseIdentity={canUseIdentity}
          identity={identity}
          isVoting={votingRequestId === selectedRequest.id}
          isWriteLocked={isIdentityPending}
          onBack={() => {
            setSelectedRequest(null);
          }}
          onCommentCreated={(comment) => {
            const updatedRequest = {
              ...selectedRequest,
              commentCount: selectedRequest.commentCount + 1,
              comments: [...selectedRequest.comments, comment],
            };
            syncRequest(updatedRequest);
          }}
          onRequireIdentity={() => {
            setPendingIdentityAction({
              identityEpoch: identityExchangeRef.current,
              type: "comment",
            });
          }}
          onVote={() => {
            requestVote(selectedRequest);
          }}
          portal={portal}
          request={selectedRequest}
        />
      ) : null}
      {selectedUpdate ? (
        <UpdateDetail
          onBack={() => {
            setSelectedUpdate(null);
          }}
          update={selectedUpdate}
        />
      ) : null}
      {isComposing ? (
        <FeedbackComposer
          canUseIdentity={canUseIdentity}
          identity={composerIdentity}
          isWriteLocked={isIdentityPending}
          onBack={() => {
            setIsComposing(false);
            setComposerIdentity(null);
          }}
          onCreated={(result) => {
            setRequests((current) => [result.request, ...current]);
            setIsComposing(false);
            setComposerIdentity(null);
            setSubmissionSuccess(result);
          }}
          onRequireIdentity={() => {
            setPendingIdentityAction({
              identityEpoch: identityExchangeRef.current,
              type: "submit",
            });
          }}
          portal={portal}
        />
      ) : null}
      {submissionSuccess ? (
        <SubmissionSuccess
          onView={() => {
            setSelectedRequest(submissionSuccess.request);
            setSubmissionSuccess(null);
          }}
          request={submissionSuccess.request}
        />
      ) : null}
      {pendingIdentityAction ? (
        <IdentityGate
          onBack={() => {
            setPendingIdentityAction(null);
          }}
          onVerified={handleVerified}
          portal={portal}
        />
      ) : null}
    </div>
  );
};
