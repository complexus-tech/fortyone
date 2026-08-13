"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import {
  ArrowLeft2Icon,
  ArrowUpDownIcon,
  BellIcon,
  CheckIcon,
  CloseIcon,
  EditIcon,
  FilterIcon,
  HomeIcon,
  ImageIcon,
  RequestsIcon,
  RoadmapIcon,
  SearchIcon,
  ThumbsDownIcon,
  ThumbsUpIcon,
  UpdatesIcon,
} from "icons";
import { Avatar, Box, Button, Flex, Input, Switch, Text } from "ui";
import { cn, getReadableTextColor } from "lib";
import type {
  PublicFeedbackListStatus,
  PublicPortal,
  PublicPortalSort,
  PublicPortalUpdate,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestComment,
  PublicRequestStatus,
} from "@/modules/public-portal/types";
import { getPublicAvatarColor } from "@/modules/public-portal/avatar-color";
import {
  requestFilters,
  requestStatusMeta,
} from "@/modules/public-portal/status";
import {
  confirmWidgetFeedbackVerificationAction,
  createWidgetFeedbackAction,
  createWidgetFeedbackCommentAction,
  exchangeWidgetIdentityAction,
  getWidgetFeedbackPageAction,
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
  type FeedbackWidgetTab,
  type FeedbackWidgetTheme,
} from "./protocol";

type WidgetRoadmap = Record<
  "completed" | "in_progress" | "planned",
  PublicRequest[]
>;

type WidgetRoadmapStatus = keyof WidgetRoadmap;

type WidgetRoadmapPagination = Record<
  WidgetRoadmapStatus,
  { hasMore: boolean; nextPage: number }
>;

type WidgetSubmissionIdentity =
  | { kind: "account" }
  | { kind: "external" | "verified_guest"; sessionToken: string };

type PendingIdentityAction =
  | { identityEpoch: number; type: "comment" }
  | {
      direction: -1 | 1;
      identityEpoch: number;
      requestId: string;
      type: "vote";
    }
  | { identityEpoch: number; type: "submit" };

const tabs = [
  { icon: HomeIcon, label: "Home", value: "home" },
  { icon: RequestsIcon, label: "Feedback", value: "feedback" },
  { icon: RoadmapIcon, label: "Roadmap", value: "roadmap" },
  { icon: UpdatesIcon, label: "Updates", value: "updates" },
] satisfies {
  icon: typeof RequestsIcon;
  label: string;
  value: FeedbackWidgetTab;
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

const mergeRequests = (current: PublicRequest[], incoming: PublicRequest[]) => {
  const requests = new Map(current.map((request) => [request.id, request]));
  incoming.forEach((request) => {
    requests.set(request.id, request);
  });
  return Array.from(requests.values());
};

const statusAccent = (status: PublicRequestStatus) => {
  if (status === "in_progress") return "bg-info";
  return requestStatusMeta[status].dotClassName;
};

const INITIAL_ROADMAP_VISIBLE_COUNT = 3;

const WidgetIconButton = ({
  children,
  className,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
  <button
    className={cn(
      "text-text-muted hover:bg-state-hover hover:text-foreground focus-visible:ring-ring inline-flex size-9 shrink-0 items-center justify-center rounded-lg transition-colors focus-visible:ring-2 focus-visible:outline-none",
      className,
    )}
    type="button"
    {...props}
  >
    {children}
  </button>
);

const WidgetBackButton = ({
  children,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) => (
  <WidgetIconButton className="bg-state-hover" {...props}>
    {children}
  </WidgetIconButton>
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
    <span className="border-border/80 text-text-muted inline-flex h-9 items-center gap-2 rounded-lg border px-3 text-[12px] font-medium">
      <span className={cn("size-2 rounded-sm", statusAccent(status))} />
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
        "border-border/80 bg-background text-text-muted hover:border-foreground/25 hover:text-foreground focus-visible:ring-ring inline-flex h-8 min-w-14 shrink-0 items-center justify-center gap-1.5 rounded-lg border px-2.5 text-[12px] font-semibold tabular-nums transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-60",
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

const DetailVoteControl = ({
  disabled,
  isPending,
  onVote,
  request,
}: {
  disabled?: boolean;
  isPending?: boolean;
  onVote: (direction: -1 | 1) => void;
  request: PublicRequest;
}) => (
  <div className="border-border/80 bg-background flex h-9 shrink-0 items-center overflow-hidden rounded-lg border">
    <button
      aria-label={
        request.viewerVote === -1 ? "Remove downvote" : "Downvote feedback"
      }
      aria-pressed={request.viewerVote === -1}
      className={cn(
        "text-text-muted hover:bg-state-hover hover:text-foreground focus-visible:ring-ring flex h-full w-9 items-center justify-center transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-60",
        { "text-foreground": request.viewerVote === -1 },
      )}
      disabled={disabled || isPending}
      onClick={() => {
        onVote(-1);
      }}
      type="button"
    >
      <ThumbsDownIcon className="h-3.5 text-current" strokeWidth={2} />
    </button>
    <span aria-hidden="true" className="bg-border/80 h-4 w-px" />
    <button
      aria-label={
        request.viewerVote === 1 ? "Remove upvote" : "Upvote feedback"
      }
      aria-pressed={request.viewerVote === 1}
      className={cn(
        "text-text-muted hover:bg-state-hover hover:text-foreground focus-visible:ring-ring flex h-full min-w-13 items-center justify-center gap-1.5 px-2.5 text-[12px] font-semibold tabular-nums transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-60",
        { "text-foreground": request.viewerVote === 1 },
      )}
      disabled={disabled || isPending}
      onClick={() => {
        onVote(1);
      }}
      type="button"
    >
      <ThumbsUpIcon className="h-3.5 text-current" strokeWidth={2} />
      {request.voteCount}
    </button>
  </div>
);

const EmptyState = ({
  body,
  icon: Icon,
  title,
}: {
  body: string;
  icon: typeof RequestsIcon;
  title: string;
}) => (
  <Flex align="center" className="px-10 py-20 text-center" direction="column">
    <Flex
      align="center"
      className="bg-surface-muted text-text-muted mb-5 size-11 rounded-lg"
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

const feedbackSortOptions: { label: string; value: PublicPortalSort }[] = [
  { label: "Top", value: "top" },
  { label: "Newest", value: "newest" },
  { label: "Oldest", value: "oldest" },
];

const feedbackStatusOptions: {
  label: string;
  value: PublicFeedbackListStatus;
}[] = [
  { label: "Active", value: "active" },
  ...requestFilters.map((status) => ({
    label: requestStatusMeta[status].label,
    value: status,
  })),
];

const WidgetFeedbackToolbar = ({
  isLoading,
  onSearchChange,
  onSortChange,
  onStatusChange,
  search,
  sort,
  status,
}: {
  isLoading: boolean;
  onSearchChange: (value: string) => void;
  onSortChange: (value: PublicPortalSort) => void;
  onStatusChange: (value: PublicFeedbackListStatus) => void;
  search: string;
  sort: PublicPortalSort;
  status: PublicFeedbackListStatus;
}) => {
  const controlsRef = useRef<HTMLDivElement | null>(null);
  const [openMenu, setOpenMenu] = useState<"sort" | "status" | null>(null);

  useEffect(() => {
    if (!openMenu) return;
    const closeOnPointerDown = (event: PointerEvent) => {
      if (!controlsRef.current?.contains(event.target as Node)) {
        setOpenMenu(null);
      }
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpenMenu(null);
    };
    document.addEventListener("pointerdown", closeOnPointerDown);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOnPointerDown);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [openMenu]);

  const menuButtonClassName =
    "text-text-muted hover:bg-state-hover hover:text-foreground focus-visible:ring-ring flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-[12px] transition-colors focus-visible:ring-2 focus-visible:outline-none";

  return (
    <Box className="bg-background sticky top-0 z-20 px-5 pt-4 pb-3">
      <Flex align="center" className="gap-2">
        <div className="border-border bg-surface focus-within:border-foreground/25 focus-within:ring-ring relative flex h-10 min-w-0 flex-1 items-center rounded-lg border px-3 focus-within:ring-2">
          <SearchIcon className="text-text-muted h-4 shrink-0" />
          <input
            aria-label="Search feedback"
            className="text-foreground placeholder:text-text-muted/65 h-full min-w-0 flex-1 border-0 bg-transparent px-2.5 text-[12px] outline-none"
            onChange={(event) => {
              onSearchChange(event.target.value);
            }}
            placeholder="Search feedback"
            type="search"
            value={search}
          />
        </div>
        <div className="relative flex shrink-0 gap-2" ref={controlsRef}>
          <WidgetIconButton
            aria-expanded={openMenu === "sort"}
            aria-haspopup="menu"
            aria-label={`Order feedback by ${feedbackSortOptions.find((option) => option.value === sort)?.label ?? sort}`}
            className={cn("border-border bg-surface border", {
              "bg-surface-elevated text-foreground": openMenu === "sort",
              "opacity-60": isLoading,
            })}
            onClick={() => {
              setOpenMenu((current) => (current === "sort" ? null : "sort"));
            }}
          >
            <ArrowUpDownIcon className="h-4 text-current" />
          </WidgetIconButton>
          <WidgetIconButton
            aria-expanded={openMenu === "status"}
            aria-haspopup="menu"
            aria-label={`Filter feedback by ${feedbackStatusOptions.find((option) => option.value === status)?.label ?? status}`}
            className={cn("border-border bg-surface border", {
              "bg-surface-elevated text-foreground": openMenu === "status",
              "opacity-60": isLoading,
            })}
            onClick={() => {
              setOpenMenu((current) =>
                current === "status" ? null : "status",
              );
            }}
          >
            <FilterIcon className="h-4 text-current" strokeWidth={2} />
          </WidgetIconButton>

          {openMenu ? (
            <Box
              aria-label={
                openMenu === "sort" ? "Feedback ordering" : "Status filters"
              }
              className="border-border bg-surface-elevated absolute top-11 right-0 z-30 min-w-44 rounded-lg border p-1.5 shadow-xl"
              role="menu"
            >
              <Text
                className="px-3 pt-1 pb-1.5 text-[10px] tracking-[0.12em] uppercase"
                color="muted"
                fontWeight="semibold"
              >
                {openMenu === "sort" ? "Order by" : "Status"}
              </Text>
              {(openMenu === "sort"
                ? feedbackSortOptions
                : feedbackStatusOptions
              ).map((option) => {
                const selected =
                  openMenu === "sort"
                    ? option.value === sort
                    : option.value === status;
                const statusMeta =
                  openMenu === "status" && option.value !== "active"
                    ? requestStatusMeta[
                        option.value as Exclude<
                          PublicFeedbackListStatus,
                          "active"
                        >
                      ]
                    : null;
                return (
                  <button
                    aria-checked={selected}
                    className={cn(menuButtonClassName, {
                      "bg-state-hover text-foreground": selected,
                    })}
                    key={option.value}
                    onClick={() => {
                      if (openMenu === "sort") {
                        onSortChange(option.value as PublicPortalSort);
                      } else {
                        onStatusChange(
                          option.value as PublicFeedbackListStatus,
                        );
                      }
                      setOpenMenu(null);
                    }}
                    role="menuitemradio"
                    type="button"
                  >
                    {statusMeta ? (
                      <span
                        className={cn(
                          "size-2 rounded-sm",
                          statusMeta.dotClassName,
                        )}
                      />
                    ) : null}
                    <span className="flex-1">{option.label}</span>
                    {selected ? (
                      <CheckIcon className="h-3.5 text-current" />
                    ) : null}
                  </button>
                );
              })}
            </Box>
          ) : null}
        </div>
      </Flex>
    </Box>
  );
};

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
        <WidgetBackButton aria-label="Go back" onClick={onBack}>
          <ArrowLeft2Icon className="h-5" />
        </WidgetBackButton>
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
                  className="border-border/70 rounded-lg border p-3"
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
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-10 w-full items-center justify-center rounded-lg px-5 text-[13px] font-semibold focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
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

const UpdateDetail = ({
  onBack,
  update,
}: {
  onBack: () => void;
  update: PublicPortalUpdate;
}) => (
  <Box className="bg-background absolute inset-0 z-20 flex min-h-0 flex-col">
    <Flex align="center" className="h-16 shrink-0 px-4">
      <WidgetBackButton aria-label="Back to updates" onClick={onBack}>
        <ArrowLeft2Icon className="h-5" />
      </WidgetBackButton>
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
        <Box className="border-border bg-surface mt-7 rounded-lg border p-4">
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
                className={cn("size-2 rounded-sm", statusAccent(item.status))}
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

const HomeSectionHeader = ({
  actionLabel,
  icon: Icon,
  onAction,
  title,
}: {
  actionLabel: string;
  icon: typeof RequestsIcon;
  onAction: () => void;
  title: string;
}) => (
  <Flex align="center" className="px-5 pt-6 pb-2" justify="between">
    <Flex align="center" className="min-w-0 gap-2">
      <Icon className="text-text-muted h-4 shrink-0" />
      <Text
        className="truncate text-[11px] tracking-[0.09em] uppercase"
        fontWeight="semibold"
      >
        {title}
      </Text>
    </Flex>
    <button
      className="text-text-muted hover:text-foreground focus-visible:ring-ring shrink-0 rounded-lg px-2 py-1 text-[10px] font-normal tracking-[0.09em] uppercase transition-colors focus-visible:ring-2 focus-visible:outline-none"
      onClick={onAction}
      type="button"
    >
      {actionLabel}
    </button>
  </Flex>
);

const RoadmapGroupHeader = ({
  count,
  label,
  status,
}: {
  count: number;
  label: string;
  status: PublicRequestStatus;
}) => (
  <Flex align="center" className="relative py-3 pr-1 pl-6" justify="between">
    <span
      className={cn(
        "ring-background absolute top-1/2 -left-1.5 size-3 -translate-y-1/2 rounded-sm ring-4",
        statusAccent(status),
      )}
    />
    <Text className="text-[15px]" fontWeight="semibold">
      {label}
    </Text>
    <Text className="text-[11px] tabular-nums" color="muted">
      {String(count).padStart(2, "0")}
    </Text>
  </Flex>
);

const HomeRoadmapRow = ({
  isVoting,
  isWriteLocked,
  onOpen,
  onVote,
  request,
}: {
  isVoting: boolean;
  isWriteLocked: boolean;
  onOpen: () => void;
  onVote: () => void;
  request: PublicRequest;
}) => {
  const status = requestStatusMeta[request.status];
  return (
    <div className="hover:bg-state-hover/35 relative grid grid-cols-[minmax(0,1fr)_auto] items-start gap-3 py-3 pr-1 pl-6 transition-colors">
      <span
        className={cn(
          "ring-background absolute top-5 -left-1 size-2.5 rounded-sm ring-4",
          statusAccent(request.status),
        )}
      />
      <button
        className="focus-visible:ring-ring min-w-0 text-left focus-visible:ring-2 focus-visible:outline-none"
        onClick={onOpen}
        type="button"
      >
        <Box className="min-w-0">
          <Text
            className="line-clamp-1 text-[13px] leading-5"
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
          <Flex align="center" className="mt-2 gap-2 text-[12px]">
            <Text className="font-medium" color="muted">
              {status.label}
            </Text>
            <span className="text-text-muted">&bull;</span>
            <Text className="truncate font-medium" color="muted">
              {request.authorName || "Anonymous"}
            </Text>
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

const WidgetHome = ({
  feedback,
  isWriteLocked,
  latestUpdate,
  onOpenFeedback,
  onOpenRoadmap,
  onOpenUpdate,
  onOpenRequest,
  onShareFeedback,
  onVote,
  portal,
  roadmap,
  votingRequestId,
}: {
  feedback: PublicRequest[];
  isWriteLocked: boolean;
  latestUpdate?: PublicPortalUpdate;
  onOpenFeedback: () => void;
  onOpenRoadmap: () => void;
  onOpenUpdate: (update: PublicPortalUpdate) => void;
  onOpenRequest: (request: PublicRequest) => void;
  onShareFeedback: () => void;
  onVote: (request: PublicRequest) => void;
  portal: PublicPortal;
  roadmap: WidgetRoadmap;
  votingRequestId: string | null;
}) => {
  const popularFeedback = feedback.slice(0, 3);
  const homeRoadmapGroups = roadmapSections
    .map((section) => ({
      ...section,
      items: roadmap[section.status].slice(0, 2),
      total: roadmap[section.status].length,
    }))
    .filter((section) => section.items.length > 0);

  return (
    <Box className="pb-4">
      <section aria-labelledby="widget-home-feedback">
        <div id="widget-home-feedback">
          <HomeSectionHeader
            actionLabel="See all"
            icon={RequestsIcon}
            onAction={onOpenFeedback}
            title="Popular feedback"
          />
        </div>
        {popularFeedback.length > 0 ? (
          popularFeedback.map((request) => (
            <FeedbackRow
              isVoting={votingRequestId === request.id}
              isWriteLocked={isWriteLocked}
              key={request.id}
              onOpen={() => {
                onOpenRequest(request);
              }}
              onVote={() => {
                onVote(request);
              }}
              portal={portal}
              request={request}
            />
          ))
        ) : (
          <Text className="block px-5 py-5 text-[12px]" color="muted">
            Feedback shared by customers will appear here.
          </Text>
        )}
      </section>

      <Box className="border-border/60 border-y p-5">
        <button
          className="border-border bg-state-hover/40 hover:bg-state-hover/55 focus-visible:ring-ring flex w-full items-center gap-4 rounded-lg border p-4 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
          onClick={onShareFeedback}
          type="button"
        >
          <Box className="min-w-0 flex-1">
            <Text className="text-[15px] leading-5" fontWeight="semibold">
              Help shape what comes next
            </Text>
            <Text className="mt-1 text-[12px] leading-5" color="muted">
              Share an idea, report a problem, or suggest an improvement.
            </Text>
          </Box>
          <span className="bg-foreground text-background inline-flex h-8 shrink-0 items-center gap-1.5 rounded-xl px-3 text-[11px] font-semibold shadow-sm">
            <EditIcon className="h-3.5 text-current" />
            Share
          </span>
        </button>
      </Box>

      {latestUpdate ? (
        <button
          className="border-border/60 hover:bg-state-hover/35 focus-visible:ring-ring w-full border-b px-5 py-6 text-left transition-colors focus-visible:ring-2 focus-visible:outline-none"
          onClick={() => {
            onOpenUpdate(latestUpdate);
          }}
          type="button"
        >
          <Text
            className="text-[10px] tracking-[0.1em] uppercase"
            color="muted"
            fontWeight="semibold"
          >
            Latest update · {latestUpdate.publishedAtLabel}
          </Text>
          <Text className="mt-2 text-[18px] leading-6" fontWeight="semibold">
            {latestUpdate.title}
          </Text>
          <Text
            className="mt-2 line-clamp-3 text-[12px] leading-5"
            color="muted"
          >
            {latestUpdate.summary || latestUpdate.body}
          </Text>
        </button>
      ) : null}

      <section aria-labelledby="widget-home-roadmap">
        <div id="widget-home-roadmap">
          <HomeSectionHeader
            actionLabel="See roadmap"
            icon={RoadmapIcon}
            onAction={onOpenRoadmap}
            title="On the roadmap"
          />
        </div>
        {homeRoadmapGroups.length > 0 ? (
          <Box className="border-border/70 mr-5 ml-6 border-l border-dashed">
            {homeRoadmapGroups.map((section) => (
              <Box key={section.status}>
                <RoadmapGroupHeader
                  count={section.total}
                  label={section.label}
                  status={section.status}
                />
                {section.items.map((request) => (
                  <HomeRoadmapRow
                    isVoting={votingRequestId === request.id}
                    isWriteLocked={isWriteLocked}
                    key={request.id}
                    onOpen={() => {
                      onOpenRequest(request);
                    }}
                    onVote={() => {
                      onVote(request);
                    }}
                    request={request}
                  />
                ))}
              </Box>
            ))}
          </Box>
        ) : (
          <Text className="block px-5 py-5 text-[12px]" color="muted">
            Planned and in-progress ideas will appear here.
          </Text>
        )}
      </section>
    </Box>
  );
};

const BottomNavigation = ({
  activeTab,
  onSelect,
  showUpdates,
  unreadUpdateCount,
}: {
  activeTab: FeedbackWidgetTab;
  onSelect: (tab: FeedbackWidgetTab) => void;
  showUpdates: boolean;
  unreadUpdateCount: number;
}) => (
  <nav
    aria-label="Feedback sections"
    className={cn(
      "border-border/70 bg-background/85 supports-[backdrop-filter]:bg-background/75 grid shrink-0 border-t px-2 py-2 backdrop-blur-xl",
      showUpdates ? "grid-cols-4" : "grid-cols-3",
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
              "text-text-muted hover:text-foreground focus-visible:ring-ring flex h-12 flex-col items-center justify-center gap-1 rounded-lg text-[12px] font-medium transition-colors focus-visible:ring-2 focus-visible:outline-none",
              { "text-foreground": active },
            )}
            key={tab.value}
            onClick={() => {
              onSelect(tab.value);
            }}
            type="button"
          >
            <span className="relative">
              <Icon className="h-[18px] text-current" />
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
  roadmapPagination,
  theme,
  viewer,
}: {
  initialTab: FeedbackWidgetTab;
  instanceId: string;
  mode: FeedbackWidgetMode;
  parentOrigin: string;
  portal: PublicPortal;
  roadmap: WidgetRoadmap;
  roadmapPagination?: WidgetRoadmapPagination;
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
      ? "home"
      : initialTab,
  );
  const [requests, setRequests] = useState(portal.requests);
  const [homeRequests, setHomeRequests] = useState(portal.requests);
  const [search, setSearch] = useState("");
  const [feedbackSort, setFeedbackSort] = useState<PublicPortalSort>("top");
  const [feedbackStatus, setFeedbackStatus] =
    useState<PublicFeedbackListStatus>("active");
  const [isFeedbackLoading, setIsFeedbackLoading] = useState(false);
  const [feedbackError, setFeedbackError] = useState("");
  const feedbackQueryRef = useRef(0);
  const feedbackFiltersRef = useRef({
    search: "",
    sort: "top" as PublicPortalSort,
    status: "active" as PublicFeedbackListStatus,
  });
  const [roadmapItems, setRoadmapItems] = useState(roadmap);
  const [roadmapPageState, setRoadmapPageState] =
    useState<WidgetRoadmapPagination>(
      roadmapPagination ?? {
        completed: { hasMore: false, nextPage: 2 },
        in_progress: { hasMore: false, nextPage: 2 },
        planned: { hasMore: false, nextPage: 2 },
      },
    );
  const [visibleRoadmapCounts, setVisibleRoadmapCounts] = useState<
    Record<WidgetRoadmapStatus, number>
  >({
    completed: INITIAL_ROADMAP_VISIBLE_COUNT,
    in_progress: INITIAL_ROADMAP_VISIBLE_COUNT,
    planned: INITIAL_ROADMAP_VISIBLE_COUNT,
  });
  const [loadingRoadmapStatus, setLoadingRoadmapStatus] =
    useState<WidgetRoadmapStatus | null>(null);
  const [roadmapError, setRoadmapError] = useState("");
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

  const syncRequest = useCallback((updatedRequest: PublicRequest) => {
    setRequests((current) => replaceRequest(current, updatedRequest));
    setHomeRequests((current) => replaceRequest(current, updatedRequest));
    setRoadmapItems((current) => ({
      completed: replaceRequest(current.completed, updatedRequest),
      in_progress: replaceRequest(current.in_progress, updatedRequest),
      planned: replaceRequest(current.planned, updatedRequest),
    }));
    setSelectedRequest((current) =>
      current?.id === updatedRequest.id ? updatedRequest : current,
    );
  }, []);

  useEffect(() => {
    const nextFilters = {
      search,
      sort: feedbackSort,
      status: feedbackStatus,
    };
    const previousFilters = feedbackFiltersRef.current;
    if (
      previousFilters.search === nextFilters.search &&
      previousFilters.sort === nextFilters.sort &&
      previousFilters.status === nextFilters.status
    ) {
      return;
    }
    feedbackFiltersRef.current = nextFilters;
    const queryId = feedbackQueryRef.current + 1;
    feedbackQueryRef.current = queryId;
    const timeout = window.setTimeout(
      () => {
        setIsFeedbackLoading(true);
        setFeedbackError("");
        void getWidgetFeedbackPageAction({
          page: 1,
          portalSlug: portal.slug,
          search,
          sort: feedbackSort,
          status: feedbackStatus,
        })
          .then((response) => {
            if (feedbackQueryRef.current !== queryId) return;
            if (!response.data) {
              setFeedbackError(
                response.error?.message ?? "Unable to load feedback.",
              );
              return;
            }
            setRequests(response.data.requests);
          })
          .catch(() => {
            if (feedbackQueryRef.current === queryId) {
              setFeedbackError("Unable to load feedback.");
            }
          })
          .finally(() => {
            if (feedbackQueryRef.current === queryId) {
              setIsFeedbackLoading(false);
            }
          });
      },
      search.trim() ? 250 : 0,
    );

    return () => {
      window.clearTimeout(timeout);
    };
  }, [feedbackSort, feedbackStatus, portal.slug, search]);

  const loadMoreRoadmap = useCallback(
    async (status: WidgetRoadmapStatus) => {
      const items = roadmapItems[status];
      const visibleCount = visibleRoadmapCounts[status];
      if (visibleCount < items.length) {
        setVisibleRoadmapCounts((current) => ({
          ...current,
          [status]: Math.min(
            items.length,
            current[status] + INITIAL_ROADMAP_VISIBLE_COUNT,
          ),
        }));
        return;
      }

      const pagination = roadmapPageState[status];
      if (!pagination.hasMore || loadingRoadmapStatus) return;
      setLoadingRoadmapStatus(status);
      setRoadmapError("");
      const response = await getWidgetFeedbackPageAction({
        page: pagination.nextPage,
        portalSlug: portal.slug,
        search: "",
        sort: "newest",
        status,
      }).catch(() => null);
      setLoadingRoadmapStatus(null);
      if (!response?.data) {
        setRoadmapError("Unable to load more roadmap items.");
        return;
      }
      setRoadmapItems((current) => ({
        ...current,
        [status]: mergeRequests(current[status], response.data.requests),
      }));
      setRoadmapPageState((current) => ({
        ...current,
        [status]: {
          hasMore: response.data.hasMore,
          nextPage: response.data.nextPage,
        },
      }));
      setVisibleRoadmapCounts((current) => ({
        ...current,
        [status]: current[status] + INITIAL_ROADMAP_VISIBLE_COUNT,
      }));
    },
    [
      loadingRoadmapStatus,
      portal.slug,
      roadmapItems,
      roadmapPageState,
      visibleRoadmapCounts,
    ],
  );

  const vote = useCallback(
    async (
      request: PublicRequest,
      activeIdentity: WidgetSubmissionIdentity,
      direction: -1 | 1,
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
        vote: direction,
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
    (request: PublicRequest, direction: -1 | 1 = 1) => {
      if (identityPendingRef.current) return;
      if (!identity) {
        setPendingIdentityAction({
          direction,
          identityEpoch: identityExchangeRef.current,
          requestId: request.id,
          type: "vote",
        });
        return;
      }
      void vote(request, identity, direction);
    },
    [identity, vote],
  );

  useEffect(() => {
    const documentRoot = document.documentElement;
    const initiallyDark = documentRoot.classList.contains("dark");
    const applyTheme = (dark: boolean) => {
      documentRoot.classList.toggle("dark", dark);
    };

    if (theme !== "auto") {
      applyTheme(theme === "dark");
      return () => {
        documentRoot.classList.toggle("dark", initiallyDark);
      };
    }
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const update = () => {
      applyTheme(media.matches);
    };
    update();
    media.addEventListener("change", update);
    return () => {
      media.removeEventListener("change", update);
      documentRoot.classList.toggle("dark", initiallyDark);
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

  const openComposer = () => {
    if (identityPendingRef.current) return;
    const currentIdentity = activeIdentityRef.current;
    setComposerIdentity(currentIdentity);
    setIsComposing(true);
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
      const request =
        requests.find((item) => item.id === pendingAction.requestId) ??
        homeRequests.find((item) => item.id === pendingAction.requestId);
      if (request) {
        void vote(request, verifiedIdentity, pendingAction.direction);
      }
    }
  };
  let feedbackEmptyBody =
    "New suggestions will appear here as soon as they are shared.";
  if (feedbackStatus !== "active") {
    feedbackEmptyBody = "There is no feedback with this status yet.";
  }
  if (search) {
    feedbackEmptyBody = "Try a different phrase or share the idea yourself.";
  }
  return (
    <div
      className="bg-background/90 supports-[backdrop-filter]:bg-background/80 text-foreground relative flex h-dvh min-h-0 w-full flex-col overflow-hidden antialiased backdrop-blur-xl"
      ref={rootRef}
    >
      <Flex
        align="center"
        className="bg-background/85 supports-[backdrop-filter]:bg-background/75 relative z-10 h-16 shrink-0 gap-3 px-5 backdrop-blur-xl"
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
          className="bg-foreground text-background focus-visible:ring-ring inline-flex h-9 shrink-0 items-center gap-2 rounded-xl px-4 text-[12px] font-semibold shadow-sm focus-visible:ring-2 focus-visible:outline-none"
          disabled={isIdentityPending}
          onClick={openComposer}
          type="button"
        >
          <EditIcon className="h-3.5 text-current" />
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
        {activeTab === "home" ? (
          <WidgetHome
            feedback={homeRequests}
            isWriteLocked={isIdentityPending}
            latestUpdate={
              portal.hasPublishedUpdates ? portal.updates[0] : undefined
            }
            onOpenFeedback={() => {
              setActiveTab("feedback");
            }}
            onOpenRequest={setSelectedRequest}
            onOpenRoadmap={() => {
              setActiveTab("roadmap");
            }}
            onOpenUpdate={setSelectedUpdate}
            onShareFeedback={openComposer}
            onVote={requestVote}
            portal={portal}
            roadmap={roadmapItems}
            votingRequestId={votingRequestId}
          />
        ) : null}

        {activeTab === "feedback" ? (
          <Box>
            <WidgetFeedbackToolbar
              isLoading={isFeedbackLoading}
              onSearchChange={setSearch}
              onSortChange={setFeedbackSort}
              onStatusChange={setFeedbackStatus}
              search={search}
              sort={feedbackSort}
              status={feedbackStatus}
            />
            {feedbackError ? (
              <Text
                aria-live="polite"
                className="border-border/60 border-b px-5 py-2 text-[11px] text-red-600 dark:text-red-400"
              >
                {feedbackError}
              </Text>
            ) : null}
            <Box
              aria-busy={isFeedbackLoading}
              className={cn("transition-opacity", {
                "opacity-60": isFeedbackLoading,
              })}
            >
              {requests.length > 0 ? (
                requests.map((request) => (
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
                  body={feedbackEmptyBody}
                  icon={RequestsIcon}
                  title={search ? "No matching feedback" : "No feedback yet"}
                />
              )}
            </Box>
          </Box>
        ) : null}

        {activeTab === "roadmap" ? (
          <Box className="px-5 py-6">
            {roadmapError ? (
              <Text
                aria-live="polite"
                className="text-[11px] text-red-600 dark:text-red-400"
              >
                {roadmapError}
              </Text>
            ) : null}
            <Box className="border-border/70 ml-1.5 border-l border-dashed">
              {roadmapSections.map((section) => {
                const items = roadmapItems[section.status];
                const visibleItems = items.slice(
                  0,
                  visibleRoadmapCounts[section.status],
                );
                const hasMore =
                  visibleItems.length < items.length ||
                  roadmapPageState[section.status].hasMore;
                return (
                  <Box key={section.status}>
                    <RoadmapGroupHeader
                      count={items.length}
                      label={section.label}
                      status={section.status}
                    />
                    {items.length > 0 ? (
                      <Box>
                        {visibleItems.map((request) => (
                          <div
                            className="hover:bg-state-hover/35 relative grid w-full grid-cols-[minmax(0,1fr)_auto] items-start gap-3 py-3 pr-1 pl-6 transition-colors"
                            key={request.id}
                          >
                            <span
                              className={cn(
                                "ring-background absolute top-5 -left-[5px] size-2.5 rounded-sm ring-4",
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
                                    className="mt-1 line-clamp-2 text-[12px] leading-5"
                                    color="muted"
                                  >
                                    {request.description}
                                  </Text>
                                ) : null}
                                <Text
                                  className="mt-2 text-[12px] font-medium"
                                  color="muted"
                                >
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
                        {hasMore ? (
                          <button
                            className="text-text-muted hover:text-foreground focus-visible:ring-ring inline-flex h-9 items-center rounded-lg pr-2 pl-6 text-[12px] font-semibold transition-colors focus-visible:ring-2 focus-visible:outline-none disabled:cursor-wait disabled:opacity-60"
                            disabled={loadingRoadmapStatus !== null}
                            onClick={() => {
                              void loadMoreRoadmap(section.status);
                            }}
                            type="button"
                          >
                            {loadingRoadmapStatus === section.status
                              ? "Loading…"
                              : "Show more"}
                          </button>
                        ) : null}
                      </Box>
                    ) : (
                      <Text
                        className="py-3 pr-1 pl-6 text-[12px]"
                        color="muted"
                      >
                        Nothing here yet
                      </Text>
                    )}
                  </Box>
                );
              })}
            </Box>
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
        className="border-border/70 text-text-muted hover:text-foreground bg-background/85 supports-[backdrop-filter]:bg-background/75 focus-visible:ring-ring flex h-12 shrink-0 items-center justify-center border-t px-5 py-3 text-[12px] backdrop-blur-xl transition-colors focus-visible:ring-2 focus-visible:outline-none"
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
          onVote={(direction) => {
            requestVote(selectedRequest, direction);
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
            setHomeRequests((current) => [result.request, ...current]);
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
