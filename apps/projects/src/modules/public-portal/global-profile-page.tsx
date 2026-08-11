"use client";

import { useInfiniteQuery } from "@tanstack/react-query";
import Link from "next/link";
import { CommentIcon, RequestsIcon } from "icons";
import { Avatar, Box, Button, Flex, Text } from "ui";
import type { User } from "@/types";
import { AccountHeader } from "./account-page";
import type {
  FeedbackProfileActivity,
  FeedbackProfileActivityPage,
} from "./profile-activity";
import type { PublicPortalViewer } from "./types";
import { getCrossPortalRequestHref } from "./utils";

type ApiResponse<T> = { data: T };

const PAGE_SIZE = 20;
const ACTIVITY_DATE_FORMATTER = new Intl.DateTimeFormat("en", {
  day: "numeric",
  month: "short",
  timeZone: "UTC",
  year: "numeric",
});

const fetchActivityPage = async (page: number) => {
  const params = new URLSearchParams({
    page: String(page),
    pageSize: String(PAGE_SIZE),
  });
  const response = await fetch(`/api/profile/activity?${params.toString()}`);
  if (!response.ok) throw new Error("Unable to load feedback activity");

  const payload =
    (await response.json()) as ApiResponse<FeedbackProfileActivityPage>;
  return payload.data;
};

const formatActivityDate = (value: string) => {
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? ""
    : ACTIVITY_DATE_FORMATTER.format(date);
};

const ActivityItem = ({ activity }: { activity: FeedbackProfileActivity }) => {
  const isComment = activity.type === "comment";
  const Icon = isComment ? CommentIcon : RequestsIcon;
  const href = getCrossPortalRequestHref(
    activity.workspaceSlug,
    activity.portalSlug,
    activity.feedbackSlug,
  );

  return (
    <Link
      className="border-border/70 hover:bg-state-hover/25 group flex gap-4 border-b-[0.5px] px-1 py-5 transition-colors last:border-b-0"
      href={href}
    >
      <Flex
        align="center"
        className="bg-surface-muted text-text-muted mt-0.5 size-10 shrink-0 rounded-xl"
        justify="center"
      >
        <Icon className="h-[1.15rem] text-current" />
      </Flex>
      <Box className="min-w-0 flex-1">
        <Flex align="start" className="gap-4" justify="between">
          <Box className="min-w-0">
            <Text className="text-[0.95rem]" color="muted">
              {isComment ? "Commented on" : "Submitted feedback"}
            </Text>
            <Text
              className="mt-0.5 line-clamp-1 group-hover:underline"
              fontWeight="semibold"
            >
              {activity.feedbackTitle}
            </Text>
          </Box>
          <Text className="shrink-0 text-[0.95rem]" color="muted">
            {formatActivityDate(activity.createdAt)}
          </Text>
        </Flex>
        {activity.body ? (
          <Text className="mt-2 line-clamp-2 max-w-2xl leading-6" color="muted">
            {activity.body}
          </Text>
        ) : null}
        <Text className="mt-2 text-[0.95rem]" color="muted">
          {activity.workspaceName}
        </Text>
      </Box>
    </Link>
  );
};

export const GlobalProfilePage = ({
  initialActivity,
  profile,
  viewer,
}: {
  initialActivity: FeedbackProfileActivityPage;
  profile: User;
  viewer: PublicPortalViewer;
}) => {
  const activityQuery = useInfiniteQuery({
    queryKey: ["feedback-profile-activity"],
    queryFn: ({ pageParam }) => fetchActivityPage(pageParam),
    getNextPageParam: (lastPage) =>
      lastPage.hasMore ? lastPage.page + 1 : undefined,
    initialData: {
      pages: [initialActivity],
      pageParams: [1],
    },
    initialPageParam: 1,
  });
  const activities = activityQuery.data.pages.flatMap(
    (page) => page.activities,
  );
  const totalActivity =
    initialActivity.feedbackCount + initialActivity.commentCount;

  return (
    <Box className="bg-background min-h-dvh">
      <AccountHeader profileHref="/profile" viewer={viewer} />
      <Box className="mx-auto w-full max-w-4xl px-4 py-10 md:px-6 md:py-14">
        <Flex
          align="center"
          className="border-border bg-surface rounded-2xl border p-6"
          gap={4}
        >
          <Avatar
            className="!size-14 text-lg font-semibold"
            name={profile.fullName || profile.username}
            rounded="full"
            size="lg"
            src={profile.avatarUrl}
          />
          <Box className="min-w-0 flex-1">
            <Text as="h1" className="truncate text-2xl" fontWeight="semibold">
              {profile.fullName || profile.username}
            </Text>
            <Text className="truncate" color="muted">
              {profile.email}
            </Text>
          </Box>
          <Box className="hidden shrink-0 text-right sm:block">
            <Text className="text-xl" fontWeight="semibold">
              {totalActivity}
            </Text>
            <Text className="text-[0.95rem]" color="muted">
              {totalActivity === 1 ? "Activity" : "Activities"}
            </Text>
          </Box>
        </Flex>

        <Flex align="end" className="mt-10" justify="between">
          <Box>
            <Text as="h2" className="text-xl" fontWeight="semibold">
              Activity
            </Text>
            <Text className="mt-1" color="muted">
              Your feedback and comments across every public portal.
            </Text>
          </Box>
          <Text className="hidden text-[0.95rem] sm:block" color="muted">
            {initialActivity.feedbackCount} feedback ·{" "}
            {initialActivity.commentCount} comments
          </Text>
        </Flex>

        <Box className="border-border bg-surface mt-5 rounded-2xl border px-5">
          {activities.length > 0 ? (
            activities.map((activity) => (
              <ActivityItem
                activity={activity}
                key={`${activity.type}:${activity.id}`}
              />
            ))
          ) : (
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
                <RequestsIcon className="h-5 text-current" />
              </Flex>
              <Text fontWeight="semibold">No feedback activity yet</Text>
              <Text className="mt-1 max-w-sm" color="muted">
                Feedback and comments you share will appear here.
              </Text>
            </Flex>
          )}
        </Box>

        {activityQuery.hasNextPage ? (
          <Flex className="mt-6" justify="center">
            <Button
              color="tertiary"
              disabled={activityQuery.isFetchingNextPage}
              onClick={() => {
                void activityQuery.fetchNextPage();
              }}
              variant="outline"
            >
              {activityQuery.isFetchingNextPage ? "Loading…" : "Load more"}
            </Button>
          </Flex>
        ) : null}
        {activityQuery.isError ? (
          <Text className="mt-6 text-center" color="danger">
            Unable to load more activity. Please try again.
          </Text>
        ) : null}
      </Box>
    </Box>
  );
};
