"use client";

import type { ButtonHTMLAttributes } from "react";
import { useEffect, useRef, useState } from "react";
import type { RequestsIcon } from "icons";
import {
  ArrowUpDownIcon,
  CheckIcon,
  FilterIcon,
  SearchIcon,
  ThumbsDownIcon,
  ThumbsUpIcon,
} from "icons";
import { Avatar, Box, Flex, Text } from "ui";
import { cn } from "lib";
import type {
  PublicFeedbackListStatus,
  PublicPortalSort,
  PublicRequest,
  PublicRequestStatus,
} from "@/shared/feedback-widget/types";
import {
  feedbackRequestFilters as requestFilters,
  feedbackRequestStatusMeta as requestStatusMeta,
} from "@/shared/feedback-widget/status";
import { statusAccent } from "./utils";

export const WidgetIconButton = ({
  children,
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) => (
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

export const WidgetBackButton = ({
  children,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement>) => (
  <WidgetIconButton className="bg-state-hover" {...props}>
    {children}
  </WidgetIconButton>
);

export const UnreadBadge = ({ count }: { count: number }) => {
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

export const StatusBadge = ({ status }: { status: PublicRequestStatus }) => {
  const meta = requestStatusMeta[status];
  return (
    <span className="border-border/80 text-text-muted inline-flex h-9 items-center gap-2 rounded-lg border px-3 text-[12px] font-medium">
      <span className={cn("size-2 rounded-sm", statusAccent(status))} />
      {meta.label}
    </span>
  );
};

export const CompactStatusPill = ({
  status,
}: {
  status: PublicRequestStatus;
}) => {
  const meta = requestStatusMeta[status];
  return (
    <span
      className={cn(
        "inline-flex h-6 shrink-0 items-center gap-1.5 rounded-lg border px-2 text-[11px] font-medium",
        meta.badgeClassName,
      )}
    >
      <span className={cn("size-2 rounded-sm", meta.dotClassName)} />
      {meta.label}
    </span>
  );
};

export const VoteButton = ({
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

export const DetailVoteControl = ({
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

export const EmptyState = ({
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

export const WidgetFeedbackToolbar = ({
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

export const FeedbackRow = ({
  isWriteLocked,
  isVoting,
  onOpen,
  onVote,
  request,
}: {
  isWriteLocked: boolean;
  isVoting: boolean;
  onOpen: () => void;
  onVote: () => void;
  request: PublicRequest;
}) => (
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
          <span className="text-text-muted">·</span>
          <CompactStatusPill status={request.status} />
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
