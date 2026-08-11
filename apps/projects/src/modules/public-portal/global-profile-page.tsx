"use client";

import { useEffect, useRef, useState, type ReactNode } from "react";
import Link from "next/link";
import { useInfiniteQuery } from "@tanstack/react-query";
import { CommentIcon, RequestsIcon, ThumbsUpIcon } from "icons";
import { Avatar, Box, Flex, Tabs, Text } from "ui";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import type { User } from "@/types";
import { AccountHeader } from "./account-page";
import { getPublicAvatarColor } from "./avatar-color";
import type {
  FeedbackProfileActivity,
  FeedbackProfileActivityPage,
  FeedbackProfileActivityType,
} from "./profile-activity";
import { RequestStatusPill } from "./request-card";
import type { PublicPortalViewer } from "./types";
import { getCrossPortalRequestHref } from "./utils";

type ApiResponse<T> = { data: T };
type ProfileTab = "comments" | "feedback";

const PAGE_SIZE = 20;
const CONTRIBUTOR_DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  month: "long",
  timeZone: "UTC",
  year: "numeric",
});
const ACTIVITY_DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  day: "numeric",
  month: "short",
  timeZone: "UTC",
  year: "numeric",
});

const fetchActivityPage = async (
  page: number,
  type: FeedbackProfileActivityType,
) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
    type,
  });
  const response = await fetch(`/api/profile/activity?${params.toString()}`);
  if (!response.ok) throw new Error("Unable to load feedback activity");

  const payload =
    (await response.json()) as ApiResponse<FeedbackProfileActivityPage>;
  return payload.data;
};

const formatDate = (value: string, formatter: Intl.DateTimeFormat) => {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : formatter.format(date);
};

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

const ProfileStats = ({
  activity,
}: {
  activity: FeedbackProfileActivityPage;
}) => {
  const stats = [
    { label: "Feedback", value: activity.feedbackCount },
    {
      label: activity.commentCount === 1 ? "Comment" : "Comments",
      value: activity.commentCount,
    },
    {
      label:
        Math.abs(activity.voteScore) === 1 ? "Vote received" : "Votes received",
      value: activity.voteScore,
    },
  ];

  return (
    <dl className="flex flex-wrap justify-end gap-y-2 text-[0.95rem]">
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

const ProfileSidebar = ({
  activity,
  profile,
}: {
  activity: FeedbackProfileActivityPage;
  profile: User;
}) => {
  const name = profile.fullName || profile.username;
  const totalContributions = activity.feedbackCount + activity.commentCount;
  const portalLabel = activity.portalCount === 1 ? "portal" : "portals";

  return (
    <aside className="space-y-8 md:sticky md:top-24 md:self-start">
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
            name={name}
            rounded="full"
            size="sm"
            src={profile.avatarUrl}
            style={{ backgroundColor: getPublicAvatarColor(name) }}
          />
          <Box className="min-w-0">
            <Text className="truncate" fontWeight="semibold">
              {name}
            </Text>
            <Text className="truncate text-[0.95rem]" color="muted">
              Contributing across {activity.portalCount} {portalLabel}
            </Text>
          </Box>
        </Flex>
        <Text className="mt-4 leading-6" color="muted">
          Shares feedback and joins discussions that help shape what teams build
          next.
        </Text>
        <Box className="border-border/70 mt-5 border-t-[0.5px] pt-4">
          <Flex align="center" justify="between">
            <Text color="muted">Contributor since</Text>
            <Text fontWeight="medium">
              {formatDate(profile.createdAt, CONTRIBUTOR_DATE_FORMATTER)}
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

const FeedbackItem = ({
  activity,
  profile,
}: {
  activity: FeedbackProfileActivity;
  profile: User;
}) => {
  const name = profile.fullName || profile.username;
  const href = getCrossPortalRequestHref(
    activity.workspaceSlug,
    activity.portalSlug,
    activity.feedbackSlug,
  );

  return (
    <Box className="hover:bg-state-hover/25 group transition-colors">
      <Box className="border-border/70 border-b-[0.5px] py-5">
        <Flex align="start" className="gap-3">
          <Link className="flex min-w-0 flex-1 gap-4" href={href}>
            <Avatar
              className="mt-0.5"
              name={name}
              size="sm"
              src={profile.avatarUrl}
              style={{ backgroundColor: getPublicAvatarColor(name) }}
            />
            <Box className="min-w-0 flex-1">
              <Flex align="center" className="min-w-0 flex-wrap gap-1.5">
                <Text fontWeight="semibold">{name}</Text>
                {activity.boardName ? (
                  <>
                    <Text color="muted">in</Text>
                    <Text color="muted">{activity.boardName}</Text>
                  </>
                ) : null}
                <Text color="muted">· {activity.workspaceName}</Text>
              </Flex>
              <Text
                className="mt-1.5 line-clamp-1 text-[1.08rem] group-hover:opacity-90"
                fontWeight="semibold"
              >
                {activity.feedbackTitle}
              </Text>
              {activity.body ? (
                <Text className="mt-1.5 line-clamp-2" color="muted">
                  {activity.body}
                </Text>
              ) : null}
              <Flex align="center" className="mt-4 gap-2">
                {activity.status ? (
                  <RequestStatusPill status={activity.status} />
                ) : null}
                {activity.commentCount > 0 ? (
                  <Flex
                    align="center"
                    aria-label={`${activity.commentCount} comments`}
                    className="text-text-muted gap-1"
                  >
                    <CommentIcon className="h-4" />
                    <span>{activity.commentCount}</span>
                  </Flex>
                ) : null}
              </Flex>
            </Box>
          </Link>
          <Flex
            align="center"
            aria-label={`${activity.voteCount} votes`}
            className="text-text-muted h-7 shrink-0 gap-1 px-1.5"
          >
            <ThumbsUpIcon className="h-3.5" strokeWidth={2} />
            <span>{activity.voteCount}</span>
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};

const CommentItem = ({ activity }: { activity: FeedbackProfileActivity }) => (
  <Link
    className="hover:bg-state-hover/25 border-border/70 block border-b-[0.5px] py-5 transition-colors"
    href={getCrossPortalRequestHref(
      activity.workspaceSlug,
      activity.portalSlug,
      activity.feedbackSlug,
    )}
  >
    <Text className="text-[0.95rem]" color="muted">
      Commented on{" "}
      <span className="text-text-primary font-medium">
        {activity.feedbackTitle}
      </span>
    </Text>
    <Text className="mt-2 line-clamp-4 max-w-2xl leading-6">
      {activity.body}
    </Text>
    <Text className="mt-2 text-[0.95rem]" color="muted">
      {formatDate(activity.createdAt, ACTIVITY_DATE_FORMATTER)} ·{" "}
      {activity.workspaceName}
    </Text>
  </Link>
);

export const GlobalProfilePage = ({
  initialActivity,
  profile,
  viewer,
}: {
  initialActivity: FeedbackProfileActivityPage;
  profile: User;
  viewer: PublicPortalViewer;
}) => {
  const feedbackSentinelRef = useRef<HTMLDivElement | null>(null);
  const commentsSentinelRef = useRef<HTMLDivElement | null>(null);
  const [activeTab, setActiveTab] = useState<ProfileTab>("feedback");
  const feedbackQuery = useInfiniteQuery({
    queryKey: ["feedback-profile-activity", "feedback"],
    queryFn: ({ pageParam }) => fetchActivityPage(pageParam, "feedback"),
    getNextPageParam: (lastPage) =>
      lastPage.hasMore ? lastPage.page + 1 : undefined,
    initialData: {
      pages: [initialActivity],
      pageParams: [1],
    },
    initialPageParam: 1,
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
  const commentsQuery = useInfiniteQuery({
    queryKey: ["feedback-profile-activity", "comment"],
    queryFn: ({ pageParam }) => fetchActivityPage(pageParam, "comment"),
    enabled: activeTab === "comments",
    getNextPageParam: (lastPage) =>
      lastPage.hasMore ? lastPage.page + 1 : undefined,
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
  const feedback = feedbackData.pages.flatMap((page) => page.activities);
  const comments = (commentsData?.pages ?? []).flatMap(
    (page) => page.activities,
  );

  useEffect(() => {
    const restoreTab = () => {
      const tab = new URLSearchParams(window.location.search).get("tab");
      setActiveTab(tab === "comments" ? "comments" : "feedback");
    };

    restoreTab();
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
    const nextTab: ProfileTab = value === "comments" ? "comments" : "feedback";
    const url = new URL(window.location.href);

    if (nextTab === "comments") {
      url.searchParams.set("tab", "comments");
    } else {
      url.searchParams.delete("tab");
    }
    window.history.pushState(window.history.state, "", url);
    setActiveTab(nextTab);
  };
  const name = profile.fullName || profile.username;
  let commentsContent: ReactNode = (
    <EmptyContributionState
      description="Comments you share will appear here."
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
    commentsContent = comments.map((activity) => (
      <CommentItem activity={activity} key={activity.id} />
    ));
  }

  return (
    <Box className="bg-background min-h-dvh">
      <AccountHeader profileHref="/profile" viewer={viewer} />
      <Box className="mx-auto grid w-full max-w-[78rem] gap-10 px-4 pt-8 md:grid-cols-[minmax(0,1fr)_19rem] md:px-6 md:pt-10">
        <Flex className="min-h-0" direction="column">
          <Flex align="center" className="shrink-0 gap-4">
            <Avatar
              className="!size-16 text-xl"
              name={name}
              rounded="full"
              size="lg"
              src={profile.avatarUrl}
              style={{ backgroundColor: getPublicAvatarColor(name) }}
            />
            <Box className="min-w-0">
              <Text as="h1" className="truncate text-2xl" fontWeight="semibold">
                {name}
              </Text>
              <Text className="mt-1" color="muted">
                Feedback contributor across FortyOne
              </Text>
            </Box>
          </Flex>

          <Tabs
            className="mt-8 flex min-h-0 flex-1 flex-col"
            onValueChange={changeTab}
            value={activeTab}
          >
            <Box className="border-border/60 bg-background/85 supports-[backdrop-filter]:bg-background/70 sticky top-16 z-10 shrink-0 border-b py-3 backdrop-blur-xl">
              <Flex
                align="center"
                className="flex-wrap gap-x-6 gap-y-3"
                justify="between"
              >
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
                <ProfileStats activity={initialActivity} />
              </Flex>
            </Box>
            <Tabs.Panel className="min-h-0 md:flex-1" value="feedback">
              {feedback.length > 0 ? (
                feedback.map((activity) => (
                  <FeedbackItem
                    activity={activity}
                    key={activity.id}
                    profile={profile}
                  />
                ))
              ) : (
                <EmptyContributionState
                  description="Feedback you submit will appear here."
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
            <Tabs.Panel className="min-h-0 md:flex-1" value="comments">
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

        <ProfileSidebar activity={initialActivity} profile={profile} />
      </Box>
    </Box>
  );
};
