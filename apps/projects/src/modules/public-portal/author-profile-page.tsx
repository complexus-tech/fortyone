"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import Link from "next/link";
import { useInfiniteQuery } from "@tanstack/react-query";
import { CommentIcon, RequestsIcon } from "icons";
import { Avatar, Box, Flex, Tabs, Text } from "ui";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import { getPublicAvatarColor } from "./avatar-color";
import { PublicBoardPill } from "./board-pill";
import { FeedbackVoteButton } from "./feedback-controls";
import { PublicPortalShell } from "./portal-shell";
import { publicPortalKeys } from "./query-keys";
import { RequestStatusPill } from "./request-card";
import type {
  PublicContributor,
  PublicContributorComment,
  PublicContributorCommentsPage,
  PublicPortal,
  PublicPortalViewer,
  PublicRequest,
} from "./types";
import { getBoard, getRequestPath, getRequestPathBySlug } from "./utils";

type ApiResponse<T> = {
  data: T;
};

export type PublicContributorTab = "comments" | "feedback";

const PAGE_SIZE = 20;
const CONTRIBUTOR_DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  month: "long",
  timeZone: "UTC",
  year: "numeric",
});

const fetchAuthorFeedbackPage = async ({
  authorId,
  page,
  portalSlug,
}: {
  authorId: string;
  page: number;
  portalSlug: string;
}) => {
  const params = new URLSearchParams({
    authorId,
    page: String(page),
    pageSize: String(PAGE_SIZE),
    sort: "newest",
  });
  const response = await fetch(
    `/api/public-portal/${portalSlug}?${params.toString()}`,
  );

  if (!response.ok) {
    throw new Error("Unable to load this contributor's feedback");
  }

  const payload = (await response.json()) as ApiResponse<PublicPortal>;
  return payload.data;
};

const fetchAuthorCommentsPage = async ({
  authorId,
  page,
  portalSlug,
}: {
  authorId: string;
  page: number;
  portalSlug: string;
}) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
  });
  const response = await fetch(
    `/api/public-portal/${portalSlug}/contributors/${authorId}/comments?${params.toString()}`,
  );

  if (!response.ok) {
    throw new Error("Unable to load this contributor's comments");
  }

  const payload =
    (await response.json()) as ApiResponse<PublicContributorCommentsPage>;
  return payload.data;
};

const uniqueById = <T extends { id: string }>(items: T[]) =>
  Array.from(new Map(items.map((item) => [item.id, item])).values());

const AuthorFeedbackItem = ({
  portal,
  request,
}: {
  portal: PublicPortal;
  request: PublicRequest;
}) => {
  const board = getBoard(portal, request.boardId);

  return (
    <Box className="hover:bg-state-hover/25 transition-colors">
      <Box className="border-border/70 border-b-[0.5px] py-5">
        <Flex align="start" className="gap-3">
          <Link
            className="min-w-0 flex-1"
            href={getRequestPath(portal, request)}
          >
            <Flex align="center" className="min-w-0 flex-wrap gap-2">
              {board ? <PublicBoardPill board={board} /> : null}
              <Text className="text-sm" color="muted">
                {request.createdAtLabel}
              </Text>
            </Flex>
            <Text
              className="mt-3 line-clamp-2 text-[1.08rem]"
              fontWeight="semibold"
            >
              {request.title}
            </Text>
            {request.description ? (
              <Text className="mt-1.5 line-clamp-2 max-w-2xl" color="muted">
                {request.description}
              </Text>
            ) : null}
            <Flex align="center" className="mt-4 gap-2">
              <RequestStatusPill status={request.status} />
              {request.commentCount > 0 ? (
                <Flex
                  align="center"
                  aria-label={`${request.commentCount} comments`}
                  className="text-text-muted gap-1"
                >
                  <CommentIcon className="h-4" />
                  <span>{request.commentCount}</span>
                </Flex>
              ) : null}
            </Flex>
          </Link>
          <FeedbackVoteButton compact portal={portal} request={request} />
        </Flex>
      </Box>
    </Box>
  );
};

const AuthorCommentItem = ({
  comment,
  portal,
}: {
  comment: PublicContributorComment;
  portal: PublicPortal;
}) => (
  <Link
    className="hover:bg-state-hover/25 border-border/70 block border-b-[0.5px] py-5 transition-colors"
    href={getRequestPathBySlug(portal, comment.feedback.slug)}
  >
    <Text className="text-sm" color="muted">
      Commented on{" "}
      <span className="text-text-primary font-medium">
        {comment.feedback.title}
      </span>
    </Text>
    <Text className="mt-2 line-clamp-4 max-w-2xl leading-6">
      {comment.body}
    </Text>
    <Text className="mt-2 text-sm" color="muted">
      {comment.createdAtLabel}
    </Text>
  </Link>
);

const EmptyContributionState = ({
  description,
  icon,
  title,
}: {
  description: string;
  icon: ReactNode;
  title: string;
}) => (
  <Flex
    align="center"
    className="min-h-72 text-center"
    direction="column"
    justify="center"
  >
    <Flex
      align="center"
      className="bg-surface-muted text-text-muted mb-4 size-12 rounded-xl"
      justify="center"
    >
      {icon}
    </Flex>
    <Text className="text-[1.05rem]" fontWeight="semibold">
      {title}
    </Text>
    <Text className="mt-1 max-w-sm" color="muted">
      {description}
    </Text>
  </Flex>
);

const formatContributorSince = (joinedAt: string) => {
  const date = new Date(joinedAt);
  if (Number.isNaN(date.getTime())) return "—";

  return CONTRIBUTOR_DATE_FORMATTER.format(date);
};

const ContributorStats = ({
  contributor,
}: {
  contributor: PublicContributor;
}) => {
  const voteLabel =
    Math.abs(contributor.stats.voteScore) === 1
      ? "Vote received"
      : "Votes received";
  const stats = [
    { label: "Feedback", value: contributor.stats.feedbackCount },
    {
      label: contributor.stats.commentCount === 1 ? "Comment" : "Comments",
      value: contributor.stats.commentCount,
    },
    { label: voteLabel, value: contributor.stats.voteScore },
  ];

  return (
    <dl className="mt-6 flex flex-wrap gap-y-2 text-[0.95rem]">
      {stats.map((stat) => (
        <div
          className="border-border/70 flex items-baseline gap-1.5 border-l px-4 first:border-l-0 first:pl-0"
          key={stat.label}
        >
          <dd className="text-text-primary font-semibold">{stat.value}</dd>
          <dt className="text-text-muted">{stat.label}</dt>
        </div>
      ))}
    </dl>
  );
};

const ContributorSidebar = ({
  contributor,
  portal,
}: {
  contributor: PublicContributor;
  portal: PublicPortal;
}) => {
  const totalContributions =
    contributor.stats.feedbackCount + contributor.stats.commentCount;

  return (
    <aside className="space-y-8 md:min-h-0 md:overflow-y-auto">
      <Box className="border-border bg-surface shadow-shadow/40 rounded-xl border-[0.5px] p-5 shadow-sm">
        <Text
          className="text-[0.8rem] tracking-[0.12em] uppercase"
          color="muted"
        >
          Contributor
        </Text>
        <Flex align="center" className="mt-4 gap-3">
          <Avatar
            className="!size-10"
            name={contributor.name}
            rounded="full"
            size="sm"
            src={contributor.avatarUrl}
            style={{
              backgroundColor: getPublicAvatarColor(contributor.name),
            }}
          />
          <Box className="min-w-0">
            <Text className="truncate" fontWeight="semibold">
              {contributor.name}
            </Text>
            <Text className="truncate text-sm" color="muted">
              Contributing to {portal.workspace.name}
            </Text>
          </Box>
        </Flex>
        <Text className="mt-4 leading-6" color="muted">
          Shares feedback and joins discussions that help shape what the team
          builds next.
        </Text>
        <Box className="border-border/70 mt-5 border-t-[0.5px] pt-4">
          <Flex align="center" justify="between">
            <Text color="muted">Contributor since</Text>
            <Text fontWeight="medium">
              {formatContributorSince(contributor.joinedAt)}
            </Text>
          </Flex>
          <Flex align="center" className="mt-3" justify="between">
            <Text color="muted">Total contributions</Text>
            <Text fontWeight="medium">{totalContributions}</Text>
          </Flex>
        </Box>
      </Box>
    </aside>
  );
};

export const PublicPortalAuthorProfilePage = ({
  authorId,
  contributor,
  initialComments,
  initialTab = "feedback",
  portal,
  viewer,
}: {
  authorId: string;
  contributor: PublicContributor;
  initialComments?: PublicContributorCommentsPage | null;
  initialTab?: PublicContributorTab;
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => {
  const feedbackSentinelRef = useRef<HTMLDivElement | null>(null);
  const commentsSentinelRef = useRef<HTMLDivElement | null>(null);
  const [activeTab, setActiveTab] = useState<PublicContributorTab>(initialTab);
  const feedbackQuery = useInfiniteQuery({
    queryKey: publicPortalKeys.authorFeedback(portal.slug, authorId),
    queryFn: ({ pageParam }) =>
      fetchAuthorFeedbackPage({
        authorId,
        page: pageParam,
        portalSlug: portal.slug,
      }),
    getNextPageParam: (lastPage, _pages, lastPageParam) =>
      lastPage.requestsHasMore ? lastPageParam + 1 : undefined,
    initialData: {
      pages: [portal],
      pageParams: [1],
    },
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  const commentsQuery = useInfiniteQuery({
    queryKey: publicPortalKeys.authorComments(portal.slug, authorId),
    queryFn: ({ pageParam }) =>
      fetchAuthorCommentsPage({
        authorId,
        page: pageParam,
        portalSlug: portal.slug,
      }),
    enabled: activeTab === "comments",
    getNextPageParam: (lastPage) =>
      lastPage.pagination.hasMore ? lastPage.pagination.nextPage : undefined,
    ...(initialComments
      ? {
          initialData: {
            pages: [initialComments],
            pageParams: [1],
          },
        }
      : {}),
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  const {
    data: feedbackData,
    fetchNextPage: fetchNextFeedbackPage,
    hasNextPage: hasNextFeedbackPage,
    isError: isFeedbackError,
    isFetchingNextPage: isFetchingNextFeedbackPage,
  } = feedbackQuery;
  const {
    data: commentsData,
    fetchNextPage: fetchNextCommentsPage,
    hasNextPage: hasNextCommentsPage,
    isError: isCommentsError,
    isFetchingNextPage: isFetchingNextCommentsPage,
    isPending: isCommentsPending,
  } = commentsQuery;
  const requests = uniqueById(
    feedbackData.pages.flatMap((page) => page.requests),
  );
  const comments = uniqueById(
    (commentsData?.pages ?? []).flatMap((page) => page.comments),
  );

  useEffect(() => {
    const restoreTab = () => {
      const tab = new URLSearchParams(window.location.search).get("tab");
      setActiveTab(tab === "comments" ? "comments" : "feedback");
    };

    window.addEventListener("popstate", restoreTab);
    return () => {
      window.removeEventListener("popstate", restoreTab);
    };
  }, []);

  useEffect(() => {
    const sentinel = feedbackSentinelRef.current;
    if (!sentinel || !hasNextFeedbackPage) return;

    const observer = new IntersectionObserver((entries) => {
      if (entries[0]?.isIntersecting && !isFetchingNextFeedbackPage) {
        void fetchNextFeedbackPage();
      }
    });
    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [fetchNextFeedbackPage, hasNextFeedbackPage, isFetchingNextFeedbackPage]);

  useEffect(() => {
    const sentinel = commentsSentinelRef.current;
    if (activeTab !== "comments" || !sentinel || !hasNextCommentsPage) {
      return;
    }

    const observer = new IntersectionObserver((entries) => {
      if (entries[0]?.isIntersecting && !isFetchingNextCommentsPage) {
        void fetchNextCommentsPage();
      }
    });
    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [
    activeTab,
    fetchNextCommentsPage,
    hasNextCommentsPage,
    isFetchingNextCommentsPage,
  ]);

  const changeTab = (value: string) => {
    const nextTab: PublicContributorTab =
      value === "comments" ? "comments" : "feedback";
    const url = new URL(window.location.href);

    if (nextTab === "comments") {
      url.searchParams.set("tab", "comments");
    } else {
      url.searchParams.delete("tab");
    }
    window.history.pushState(window.history.state, "", url);
    setActiveTab(nextTab);
  };

  let commentsContent: ReactNode = (
    <EmptyContributionState
      description="Comments shared by this contributor will appear here."
      icon={<CommentIcon className="h-5 text-current" />}
      title="No comments yet"
    />
  );
  if (isCommentsPending) {
    commentsContent = (
      <Text className="py-10 text-center" color="muted">
        Loading comments…
      </Text>
    );
  } else if (comments.length > 0) {
    commentsContent = comments.map((comment) => (
      <AuthorCommentItem comment={comment} key={comment.id} portal={portal} />
    ));
  }

  return (
    <PublicPortalShell activeTab="feedback" portal={portal} viewer={viewer}>
      <Box className="mx-auto grid w-full max-w-[78rem] gap-10 px-4 pt-8 md:h-full md:min-h-0 md:grid-cols-[minmax(0,1fr)_19rem] md:overflow-hidden md:px-6 md:pt-10">
        <Flex className="min-h-0 md:h-full" direction="column">
          <Box className="shrink-0">
            <Flex align="center" className="gap-4">
              <Avatar
                className="!size-16 text-xl"
                name={contributor.name}
                rounded="full"
                size="lg"
                src={contributor.avatarUrl}
                style={{
                  backgroundColor: getPublicAvatarColor(contributor.name),
                }}
              />
              <Box className="min-w-0">
                <Text
                  as="h1"
                  className="truncate text-2xl"
                  fontWeight="semibold"
                >
                  {contributor.name}
                </Text>
                <Text className="mt-1" color="muted">
                  Feedback contributor at {portal.workspace.name}
                </Text>
              </Box>
            </Flex>
            <ContributorStats contributor={contributor} />
          </Box>

          <Tabs
            className="mt-8 flex min-h-0 flex-1 flex-col"
            onValueChange={changeTab}
            value={activeTab}
          >
            <Box className="border-border/60 shrink-0 border-b pb-3">
              <Tabs.List className="mx-0 shrink-0 md:mx-0">
                <Tabs.Tab
                  leftIcon={<RequestsIcon className="h-4 text-current" />}
                  value="feedback"
                >
                  Feedback
                </Tabs.Tab>
                <Tabs.Tab
                  leftIcon={<CommentIcon className="h-4 text-current" />}
                  value="comments"
                >
                  Comments
                </Tabs.Tab>
              </Tabs.List>
            </Box>
            <Tabs.Panel
              className="hide-scrollbar min-h-0 pt-3 md:flex-1 md:overflow-y-auto"
              value="feedback"
            >
              {requests.length > 0 ? (
                requests.map((request) => (
                  <AuthorFeedbackItem
                    key={request.id}
                    portal={portal}
                    request={request}
                  />
                ))
              ) : (
                <EmptyContributionState
                  description="Feedback submitted by this contributor will appear here."
                  icon={<RequestsIcon className="h-5 text-current" />}
                  title="No feedback yet"
                />
              )}
              <div ref={feedbackSentinelRef} />
              {isFetchingNextFeedbackPage ? (
                <Text className="py-6 text-center" color="muted">
                  Loading more feedback…
                </Text>
              ) : null}
              {isFeedbackError ? (
                <Text className="py-6 text-center" color="muted">
                  Unable to load more feedback right now.
                </Text>
              ) : null}
            </Tabs.Panel>
            <Tabs.Panel
              className="hide-scrollbar min-h-0 pt-3 md:flex-1 md:overflow-y-auto"
              value="comments"
            >
              {commentsContent}
              <div ref={commentsSentinelRef} />
              {isFetchingNextCommentsPage ? (
                <Text className="py-6 text-center" color="muted">
                  Loading more comments…
                </Text>
              ) : null}
              {isCommentsError ? (
                <Text className="py-6 text-center" color="muted">
                  Unable to load comments right now.
                </Text>
              ) : null}
            </Tabs.Panel>
          </Tabs>
        </Flex>

        <ContributorSidebar contributor={contributor} portal={portal} />
      </Box>
    </PublicPortalShell>
  );
};
