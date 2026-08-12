"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import {
  ArrowLeft2Icon,
  CloseIcon,
  CommentIcon,
  ExternalLinkIcon,
  PlusIcon,
  RequestsIcon,
  RoadmapIcon,
  SearchIcon,
  ThumbsUpIcon,
} from "icons";
import { Avatar, Box, Button, Flex, Input, Text } from "ui";
import { cn, getReadableTextColor } from "lib";
import type {
  PublicPortal,
  PublicPortalTab,
  PublicPortalViewer,
  PublicRequest,
  PublicRequestStatus,
} from "@/modules/public-portal/types";
import { requestStatusMeta } from "@/modules/public-portal/status";
import { createWidgetFeedbackAction } from "./actions";
import type { CreateWidgetFeedbackResult } from "./actions";
import {
  FEEDBACK_WIDGET_CHANNEL,
  FEEDBACK_WIDGET_VERSION,
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

const tabs = [
  { icon: RequestsIcon, label: "Feedback", value: "feedback" },
  { icon: RoadmapIcon, label: "Roadmap", value: "roadmap" },
] satisfies {
  icon: typeof RequestsIcon;
  label: string;
  value: PublicPortalTab;
}[];

const roadmapSections = [
  {
    description: "Committed and queued",
    label: "Planned",
    status: "planned",
  },
  {
    description: "Actively being delivered",
    label: "In progress",
    status: "in_progress",
  },
  {
    description: "Recently completed",
    label: "Done",
    status: "completed",
  },
] as const;

const StatusPill = ({ status }: { status: PublicRequestStatus }) => {
  const meta = requestStatusMeta[status];
  return (
    <span
      className={cn(
        "inline-flex h-6 items-center gap-1.5 rounded-lg border px-2 text-[11px] font-semibold",
        meta.badgeClassName,
      )}
    >
      <span className={cn("size-1.5 rounded-sm", meta.dotClassName)} />
      {meta.label}
    </span>
  );
};

const FeedbackCard = ({
  onOpen,
  portal,
  request,
}: {
  onOpen: () => void;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = portal.boards.find(
    (candidate) => candidate.id === request.boardId,
  );
  return (
    <button
      className="border-border/60 hover:bg-state-hover/40 focus-visible:ring-ring group w-full border-b px-4 py-4 text-left transition-colors last:border-b-0 focus-visible:ring-2 focus-visible:outline-none"
      onClick={onOpen}
      type="button"
    >
      <Flex align="start" className="gap-3">
        <Avatar
          className="mt-0.5 shrink-0"
          name={request.authorName || "Anonymous"}
          rounded="md"
          size="sm"
          src={request.authorAvatar}
        />
        <Box className="min-w-0 flex-1">
          <Flex align="center" className="min-w-0 gap-1.5">
            <Text className="truncate text-[12px]" fontWeight="semibold">
              {request.authorName || "Anonymous"}
            </Text>
            {board ? (
              <Text className="truncate text-[11px]" color="muted">
                in {board.name}
              </Text>
            ) : null}
          </Flex>
          <Text
            className="mt-1.5 line-clamp-2 text-[14px] leading-5"
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
          <Flex align="center" className="mt-3 gap-2.5">
            <StatusPill status={request.status} />
            <Flex align="center" className="text-text-muted gap-1 text-[11px]">
              <ThumbsUpIcon className="h-3.5" />
              {request.voteCount}
            </Flex>
            {request.commentCount > 0 ? (
              <Flex
                align="center"
                className="text-text-muted gap-1 text-[11px]"
              >
                <CommentIcon className="h-3.5" />
                {request.commentCount}
              </Flex>
            ) : null}
          </Flex>
        </Box>
      </Flex>
    </button>
  );
};

const EmptyState = ({
  body,
  icon: Icon,
  title,
}: {
  body: string;
  icon: typeof RequestsIcon;
  title: string;
}) => (
  <Flex align="center" className="px-8 py-16 text-center" direction="column">
    <Flex
      align="center"
      className="bg-surface-muted text-text-muted mb-4 size-12 rounded-2xl"
      justify="center"
    >
      <Icon className="h-5" />
    </Flex>
    <Text className="text-[14px]" fontWeight="semibold">
      {title}
    </Text>
    <Text className="mt-1.5 max-w-64 text-[12px] leading-5" color="muted">
      {body}
    </Text>
  </Flex>
);

const FeedbackComposer = ({
  isAnonymous,
  onClose,
  onCreated,
  portal,
}: {
  isAnonymous: boolean;
  onClose: () => void;
  onCreated: (result: CreateWidgetFeedbackResult) => void;
  portal: PublicPortal;
}) => {
  const [boardId, setBoardId] = useState(
    portal.boards.length === 1 ? portal.boards[0]?.id ?? "" : "",
  );
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  const submit = async () => {
    if (!boardId || !title.trim() || isSubmitting) return;
    setIsSubmitting(true);
    setError("");
    const response = await createWidgetFeedbackAction({
      boardId,
      description: description.trim(),
      isAnonymous,
      portalSlug: portal.slug,
      title: title.trim(),
    });
    setIsSubmitting(false);

    if (response.error?.message || !response.data) {
      setError(response.error?.message ?? "Unable to submit feedback");
      return;
    }
    if (response.data.isAnonymous !== isAnonymous) {
      setError("The submission privacy setting could not be confirmed.");
      return;
    }
    onCreated(response.data);
  };

  return (
    <Box className="bg-background absolute inset-0 z-30 flex min-h-0 flex-col">
      <Flex
        align="center"
        className="border-border/60 h-14 shrink-0 border-b px-3"
        gap={2}
      >
        <Button
          aria-label="Back to feedback"
          asIcon
          color="tertiary"
          onClick={onClose}
          size="sm"
          variant="naked"
        >
          <ArrowLeft2Icon className="h-4" />
        </Button>
        <Text className="text-[14px]" fontWeight="semibold">
          Share feedback
        </Text>
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
        <Box className="space-y-5">
          {portal.boards.length > 1 ? (
            <label className="block space-y-2">
              <Text
                as="span"
                className="block text-[12px]"
                fontWeight="semibold"
              >
                Board
              </Text>
              <select
                className="border-border bg-surface ring-ring h-10 w-full rounded-xl border px-3 text-[13px] outline-none focus-visible:ring-2"
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
          <label className="block space-y-2" htmlFor="feedback-widget-title">
            <Text as="span" className="block text-[12px]" fontWeight="semibold">
              What could be better?
            </Text>
            <Input
              autoFocus
              id="feedback-widget-title"
              maxLength={200}
              onChange={(event) => {
                setTitle(event.target.value);
              }}
              placeholder="A short, specific title"
              value={title}
            />
          </label>
          <label className="block space-y-2">
            <Text as="span" className="block text-[12px]" fontWeight="semibold">
              More context
            </Text>
            <textarea
              className="border-border bg-surface ring-ring min-h-36 w-full resize-y rounded-xl border p-3 text-[13px] leading-5 outline-none placeholder:text-[var(--color-text-muted)] focus-visible:ring-2"
              maxLength={5000}
              onChange={(event) => {
                setDescription(event.target.value);
              }}
              placeholder="Describe the problem, context, or expected outcome…"
              value={description}
            />
          </label>
          {isAnonymous ? (
            <Box className="border-border/60 bg-surface-muted/45 rounded-xl border p-3">
              <Text className="text-[11px] leading-5" color="muted">
                This will be posted as Anonymous. You won&apos;t receive
                personal notifications, but you can open the public feedback
                after submitting to check its status.
              </Text>
            </Box>
          ) : null}
          {error ? (
            <Text className="text-[12px] text-red-600 dark:text-red-400">
              {error}
            </Text>
          ) : null}
        </Box>
      </Box>
      <Flex
        align="center"
        className="border-border/60 shrink-0 border-t px-4 py-3"
        gap={2}
        justify="end"
      >
        <Button color="tertiary" onClick={onClose} size="sm">
          Cancel
        </Button>
        <Button
          color="invert"
          disabled={!boardId || !title.trim() || isSubmitting}
          onClick={() => {
            void submit();
          }}
          size="sm"
        >
          {isSubmitting ? "Submitting…" : "Submit feedback"}
        </Button>
      </Flex>
    </Box>
  );
};

const SubmissionSuccess = ({
  onDone,
  request,
}: {
  onDone: () => void;
  request: PublicRequest;
}) => (
  <Box className="bg-background absolute inset-0 z-40 flex flex-col">
    <Flex
      align="center"
      className="border-border/60 h-14 shrink-0 border-b px-4"
    >
      <Text className="text-[14px]" fontWeight="semibold">
        Feedback received
      </Text>
    </Flex>
    <Flex
      align="center"
      className="min-h-0 flex-1 px-8 py-10 text-center"
      direction="column"
      justify="center"
    >
      <Flex
        align="center"
        className="mb-5 size-14 rounded-2xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
        justify="center"
      >
        <ThumbsUpIcon className="h-6" />
      </Flex>
      <Text className="text-[18px] leading-6" fontWeight="semibold">
        Thanks for helping us improve
      </Text>
      <Text className="mt-2 max-w-72 text-[12px] leading-5" color="muted">
        “{request.title}” was submitted anonymously.
      </Text>
      <Box className="border-border/60 bg-surface-muted/45 mt-6 max-w-72 rounded-xl border p-3">
        <Text className="text-[11px] leading-5" color="muted">
          Anonymous feedback cannot receive personal notifications. Open the
          public feedback now to follow its status.
        </Text>
      </Box>
      <Button className="mt-7" color="invert" onClick={onDone} size="sm">
        View feedback
      </Button>
    </Flex>
  </Box>
);

const RequestDetail = ({
  onBack,
  onOpenExternal,
  portal,
  request,
}: {
  onBack: () => void;
  onOpenExternal: () => void;
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = portal.boards.find(
    (candidate) => candidate.id === request.boardId,
  );
  return (
    <Box className="bg-background absolute inset-0 z-20 flex min-h-0 flex-col">
      <Flex
        align="center"
        className="border-border/60 h-14 shrink-0 border-b px-3"
        gap={2}
      >
        <Button
          aria-label="Back to feedback"
          asIcon
          color="tertiary"
          onClick={onBack}
          size="sm"
          variant="naked"
        >
          <ArrowLeft2Icon className="h-4" />
        </Button>
        <Text
          className="min-w-0 flex-1 truncate text-[13px]"
          fontWeight="semibold"
        >
          Feedback
        </Text>
        <Button
          aria-label="Open full feedback page"
          asIcon
          color="tertiary"
          onClick={onOpenExternal}
          size="sm"
          variant="naked"
        >
          <ExternalLinkIcon className="h-4" />
        </Button>
      </Flex>
      <Box className="min-h-0 flex-1 overflow-y-auto px-5 py-6">
        <Flex align="center" className="gap-2.5">
          <Avatar
            name={request.authorName || "Anonymous"}
            rounded="md"
            size="sm"
            src={request.authorAvatar}
          />
          <Box className="min-w-0">
            <Text className="text-[12px]" fontWeight="semibold">
              {request.authorName || "Anonymous"}
            </Text>
            <Text className="text-[11px]" color="muted">
              {request.createdAtLabel}
            </Text>
          </Box>
        </Flex>
        <Text
          as="h1"
          className="mt-6 text-[20px] leading-7"
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
        <Flex align="center" className="mt-6 flex-wrap gap-2">
          <StatusPill status={request.status} />
          {board ? (
            <span className="border-border bg-surface inline-flex h-6 items-center gap-1.5 rounded-lg border px-2 text-[11px]">
              <span
                className="size-1.5 rounded-sm"
                style={{ backgroundColor: board.color }}
              />
              {board.name}
            </span>
          ) : null}
        </Flex>
        <Box className="border-border/60 mt-7 border-t pt-5">
          <Flex align="center" className="text-text-muted gap-4 text-[12px]">
            <Flex align="center" className="gap-1.5">
              <ThumbsUpIcon className="h-4" />
              {request.voteCount} votes
            </Flex>
            <Flex align="center" className="gap-1.5">
              <CommentIcon className="h-4" />
              {request.commentCount} comments
            </Flex>
          </Flex>
          <Button
            className="mt-5 w-full justify-center"
            color="tertiary"
            onClick={onOpenExternal}
            rightIcon={<ExternalLinkIcon className="h-3.5" />}
            size="sm"
          >
            View full discussion
          </Button>
        </Box>
      </Box>
    </Box>
  );
};

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
  const [activeTab, setActiveTab] = useState(initialTab);
  const [requests, setRequests] = useState(portal.requests);
  const [search, setSearch] = useState("");
  const [selectedRequest, setSelectedRequest] = useState<PublicRequest | null>(
    null,
  );
  const [isComposing, setIsComposing] = useState(false);
  const [anonymousSuccess, setAnonymousSuccess] =
    useState<PublicRequest | null>(null);
  const [isDark, setIsDark] = useState(theme === "dark");
  const trustedParentOrigin = getTrustedWidgetOrigin(parentOrigin);
  const canSubmit =
    Boolean(viewer) || portal.participationMode === "anonymous_allowed";
  const filteredRequests = useMemo(() => {
    const value = search.trim().toLowerCase();
    if (!value) return requests;
    return requests.filter((request) =>
      `${request.title} ${request.description}`.toLowerCase().includes(value),
    );
  }, [requests, search]);

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
      ) {
        return;
      }
      if (event.data.event === "host-close") {
        setIsComposing(false);
        setSelectedRequest(null);
      }
    };
    window.addEventListener("message", handleMessage);
    return () => {
      window.removeEventListener("message", handleMessage);
    };
  }, [instanceId, trustedParentOrigin]);

  useEffect(() => {
    if (!trustedParentOrigin || mode === "inline") return;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      postFeedbackWidgetMessage("escape", instanceId, trustedParentOrigin);
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

  const openExternal = (path: string) => {
    if (!trustedParentOrigin) return;
    const href = new URL(path, window.location.origin).href;
    postFeedbackWidgetMessage(
      "open-external",
      instanceId,
      trustedParentOrigin,
      {
        href,
      },
    );
  };

  const requestClose = () => {
    if (!trustedParentOrigin) return;
    postFeedbackWidgetMessage("close", instanceId, trustedParentOrigin);
  };

  return (
    <div
      className={cn(
        "bg-background text-foreground relative flex h-dvh min-h-0 w-full flex-col overflow-hidden",
        { dark: isDark },
      )}
      ref={rootRef}
    >
      <Box className="border-border/60 bg-background/95 relative z-10 shrink-0 border-b backdrop-blur-xl">
        <Flex align="center" className="h-14 gap-2.5 px-4">
          <Avatar
            className="!size-8 text-[12px] font-bold shadow-sm"
            name={portal.workspace.name}
            rounded="lg"
            size="sm"
            src={portal.workspace.avatarUrl}
            style={{
              backgroundColor: portal.workspace.color,
              color: getReadableTextColor(portal.workspace.color),
            }}
          />
          <Box className="min-w-0 flex-1">
            <Text className="truncate text-[13px]" fontWeight="semibold">
              {portal.workspace.name}
            </Text>
            <Text
              className="text-[10px] tracking-[0.08em] uppercase"
              color="muted"
            >
              Feedback hub
            </Text>
          </Box>
          {mode !== "inline" ? (
            <Button
              aria-label="Close feedback widget"
              asIcon
              color="tertiary"
              onClick={requestClose}
              size="sm"
              variant="naked"
            >
              <CloseIcon className="h-4" />
            </Button>
          ) : null}
        </Flex>
        <nav
          aria-label="Feedback sections"
          className="grid grid-cols-3 gap-1 px-2 pb-2"
        >
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const active = activeTab === tab.value;
            return (
              <button
                aria-current={active ? "page" : undefined}
                className={cn(
                  "text-text-muted hover:text-foreground flex h-9 items-center justify-center gap-1.5 rounded-xl border border-transparent text-[11px] font-semibold transition-colors",
                  {
                    "border-border bg-surface-elevated text-foreground shadow-sm":
                      active,
                  },
                )}
                key={tab.value}
                onClick={() => {
                  setActiveTab(tab.value);
                  setSelectedRequest(null);
                }}
                type="button"
              >
                <Icon className="h-3.5" />
                {tab.label}
              </button>
            );
          })}
        </nav>
      </Box>

      <Box className="relative min-h-0 flex-1 overflow-y-auto">
        {activeTab === "feedback" ? (
          <Box>
            <Box className="border-border/60 bg-surface/35 border-b px-4 py-4">
              <Flex align="center" gap={2}>
                <Box className="min-w-0 flex-1">
                  <Text className="text-[15px]" fontWeight="semibold">
                    Help shape what comes next
                  </Text>
                  <Text className="mt-0.5 text-[11px] leading-4" color="muted">
                    Share an idea or see what others are asking for.
                  </Text>
                </Box>
                <Button
                  color="invert"
                  leftIcon={<PlusIcon className="h-3.5" />}
                  onClick={() => {
                    if (canSubmit) {
                      setIsComposing(true);
                    } else {
                      openExternal(
                        `/portal/${portal.slug}/feedback?newFeedback=true`,
                      );
                    }
                  }}
                  size="sm"
                >
                  Add feedback
                </Button>
              </Flex>
            </Box>
            <Box className="border-border/60 sticky top-0 z-10 border-b bg-[var(--color-background)] px-4 py-3">
              <Input
                className="h-9"
                leftIcon={<SearchIcon className="h-3.5" />}
                onChange={(event) => {
                  setSearch(event.target.value);
                }}
                placeholder="Search feedback"
                type="search"
                value={search}
                variant="solid"
              />
            </Box>
            {filteredRequests.length > 0 ? (
              filteredRequests.map((request) => (
                <FeedbackCard
                  key={request.id}
                  onOpen={() => {
                    setSelectedRequest(request);
                  }}
                  portal={portal}
                  request={request}
                />
              ))
            ) : (
              <EmptyState
                body={
                  search
                    ? "Try a different phrase or be the first to share this idea."
                    : "New suggestions will appear here as soon as they are shared."
                }
                icon={RequestsIcon}
                title={search ? "No matching feedback" : "No feedback yet"}
              />
            )}
          </Box>
        ) : null}

        {activeTab === "roadmap" ? (
          <Box className="space-y-6 px-4 py-5">
            <Box>
              <Text className="text-[15px]" fontWeight="semibold">
                Roadmap
              </Text>
              <Text className="mt-1 text-[11px] leading-4" color="muted">
                A clear view of what is planned, underway, and shipped.
              </Text>
            </Box>
            {roadmapSections.map((section) => {
              const items = roadmap[section.status];
              return (
                <Box key={section.status}>
                  <Flex align="center" justify="between">
                    <Box>
                      <Text className="text-[12px]" fontWeight="semibold">
                        {section.label}
                      </Text>
                      <Text className="text-[10px]" color="muted">
                        {section.description}
                      </Text>
                    </Box>
                    <span className="bg-surface-muted text-text-muted rounded-lg px-2 py-1 text-[10px] tabular-nums">
                      {items.length}
                    </span>
                  </Flex>
                  <Box className="mt-2 space-y-2">
                    {items.length > 0 ? (
                      items.map((request) => (
                        <button
                          className="border-border/60 bg-surface hover:bg-state-hover/40 w-full rounded-xl border p-3 text-left transition-colors"
                          key={request.id}
                          onClick={() => {
                            setSelectedRequest(request);
                          }}
                          type="button"
                        >
                          <Text
                            className="line-clamp-2 text-[12px] leading-5"
                            fontWeight="semibold"
                          >
                            {request.title}
                          </Text>
                          <Flex
                            align="center"
                            className="mt-2 text-[10px]"
                            gap={1}
                          >
                            <ThumbsUpIcon className="h-3" />
                            <span className="text-text-muted">
                              {request.voteCount}
                            </span>
                          </Flex>
                        </button>
                      ))
                    ) : (
                      <Box className="border-border/60 rounded-xl border border-dashed px-3 py-6 text-center">
                        <Text className="text-[11px]" color="muted">
                          Nothing here yet
                        </Text>
                      </Box>
                    )}
                  </Box>
                </Box>
              );
            })}
          </Box>
        ) : null}
      </Box>

      <Flex
        align="center"
        className="border-border/60 bg-background/95 shrink-0 border-t px-4 py-2 text-[10px] backdrop-blur"
        justify="center"
      >
        <Text color="muted">Powered by FortyOne</Text>
      </Flex>

      {selectedRequest ? (
        <RequestDetail
          onBack={() => {
            setSelectedRequest(null);
          }}
          onOpenExternal={() => {
            openExternal(
              `/portal/${portal.slug}/feedback/${selectedRequest.slug}`,
            );
          }}
          portal={portal}
          request={selectedRequest}
        />
      ) : null}
      {isComposing ? (
        <FeedbackComposer
          isAnonymous={!viewer}
          onClose={() => {
            setIsComposing(false);
          }}
          onCreated={(result) => {
            setRequests((current) => [result.request, ...current]);
            setIsComposing(false);
            if (result.isAnonymous) {
              setAnonymousSuccess(result.request);
              return;
            }
            setSelectedRequest(result.request);
          }}
          portal={portal}
        />
      ) : null}
      {anonymousSuccess ? (
        <SubmissionSuccess
          onDone={() => {
            setSelectedRequest(anonymousSuccess);
            setAnonymousSuccess(null);
          }}
          request={anonymousSuccess}
        />
      ) : null}
    </div>
  );
};

export const feedbackWidgetProtocol = {
  channel: FEEDBACK_WIDGET_CHANNEL,
  version: FEEDBACK_WIDGET_VERSION,
};
