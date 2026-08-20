"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import {
  AiIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  MoreHorizontalIcon,
  RefreshIcon,
  TimeScheduleIcon,
} from "icons";
import { Box, Flex, Menu, Text } from "ui";
import { ScheduleIssueDialog } from "@/components/ui/story/schedule-issue-dialog";
import { useTerminology } from "@/hooks";
import {
  useOverrideCalendarScheduleIssue,
  useRetryCalendarScheduleIssue,
} from "@/lib/hooks/calendar/use-schedule-issues";
import { useStoryGitHubLinks } from "@/lib/hooks/github";
import {
  deriveAutoSchedulingStatus,
  getAutoSchedulingHelper,
} from "@/lib/auto-scheduling";
import { useStoryIntegrationRequestLinks } from "@/modules/integration-requests/hooks/use-story-request-links";
import type { IntegrationRequestProviderThread } from "@/modules/integration-requests/types";
import type { StoryGitHubLink } from "@/modules/settings/workspace/integrations/github/types";
import { useStoryFeedbackLinks } from "@/modules/team-feedback/hooks/use-story-feedback-links";
import type { StoryFeedbackLink } from "@/modules/team-feedback/types";
import type { AutoSchedulingStatus } from "@/modules/stories/types";
import type { DetailedStory } from "../types";
import { FeedbackBannerRow } from "./feedback-banner";
import { GitHubBannerRow } from "./github-banner";
import { IntegrationRequestBannerRow } from "./integration-request-banner";

type BannerItem = {
  id: string;
  render: (embedded: boolean) => ReactNode;
};

type MayaBannerStatus = Extract<
  AutoSchedulingStatus,
  "at_risk" | "cannot_fit" | "needs_owner" | "needs_time"
>;

const isMayaBannerStatus = (
  status: AutoSchedulingStatus,
): status is MayaBannerStatus =>
  status === "needs_owner" ||
  status === "needs_time" ||
  status === "at_risk" ||
  status === "cannot_fit";

const getGitHubBannerLinks = (links: StoryGitHubLink[]) => {
  const issueLink = links.find((link) => link.externalType === "issue");
  const pullRequestLinks = links.filter(
    (link) => link.externalType === "pull_request",
  );
  return issueLink ? [issueLink, ...pullRequestLinks] : pullRequestLinks;
};

const MayaScheduleBannerRow = ({
  embedded,
  isRetrying,
  onChooseTime,
  onRetry,
  reason,
}: {
  embedded: boolean;
  isRetrying: boolean;
  onChooseTime?: () => void;
  onRetry?: () => void;
  reason: string;
}) => (
  <Flex
    align="center"
    className={
      embedded
        ? "min-w-0 rounded-none px-4 py-3 backdrop-blur-md"
        : "border-primary/20 bg-primary/5 min-w-0 rounded-xl border px-4 py-3 backdrop-blur-md"
    }
    justify="between"
  >
    <Flex align="center" className="min-w-0 flex-1">
      <AiIcon className="text-primary h-5 shrink-0" />
      <Text
        as="span"
        className="ml-2 min-w-0 truncate"
        color="primary"
        fontWeight="medium"
        title={`Maya needs your help · ${reason}`}
      >
        Maya needs your help
        <span aria-hidden="true"> · </span>
        {reason}
      </Text>
    </Flex>
    {onRetry ? (
      <Menu>
        <Menu.Button>
          <button
            aria-label="More Maya scheduling actions"
            className="text-primary hover:text-primary/80 ml-2 shrink-0 rounded-md p-1 transition"
            type="button"
          >
            <MoreHorizontalIcon className="h-5 text-current" />
          </button>
        </Menu.Button>
        <Menu.Items align="end" className="min-w-40">
          <Menu.Group>
            <Menu.Item disabled={isRetrying} onSelect={onRetry}>
              <RefreshIcon className="h-5 w-auto" />
              {isRetrying ? "Retrying…" : "Retry"}
            </Menu.Item>
            {onChooseTime ? (
              <Menu.Item onSelect={onChooseTime}>
                <TimeScheduleIcon className="h-5 w-auto" />
                Choose time
              </Menu.Item>
            ) : null}
          </Menu.Group>
        </Menu.Items>
      </Menu>
    ) : null}
  </Flex>
);

const StoryBannerStack = ({ items }: { items: BannerItem[] }) => {
  const [activeIndex, setActiveIndex] = useState(0);
  if (items.length === 0) return null;

  const safeActiveIndex = activeIndex % items.length;
  const activeItem = items[safeActiveIndex];
  if (items.length === 1) {
    return <Box className="mb-3">{activeItem.render(false)}</Box>;
  }

  return (
    <Box aria-live="polite" className="relative mb-3 pb-1.5">
      <Box
        aria-hidden="true"
        className="border-primary/15 bg-primary/[0.03] absolute inset-x-2 bottom-0 h-4 rounded-b-xl border opacity-70 backdrop-blur-md"
      />
      <Box className="border-primary/20 bg-primary/5 relative overflow-hidden rounded-xl border backdrop-blur-md">
        {activeItem.render(true)}
        <Flex
          align="center"
          className="border-primary/15 border-t px-3 py-2"
          justify="between"
        >
          <Text className="tabular-nums" color="primary" fontSize="sm">
            {safeActiveIndex + 1} of {items.length}
          </Text>
          <Flex className="gap-1">
            <button
              aria-label="Previous story banner"
              className="text-primary hover:bg-primary/10 focus-visible:ring-ring grid size-7 place-items-center rounded-md outline-none focus-visible:ring-2"
              onClick={() => {
                setActiveIndex(
                  (safeActiveIndex - 1 + items.length) % items.length,
                );
              }}
              type="button"
            >
              <ChevronLeftIcon aria-hidden="true" className="h-4" />
            </button>
            <button
              aria-label="Next story banner"
              className="text-primary hover:bg-primary/10 focus-visible:ring-ring grid size-7 place-items-center rounded-md outline-none focus-visible:ring-2"
              onClick={() => {
                setActiveIndex((safeActiveIndex + 1) % items.length);
              }}
              type="button"
            >
              <ChevronRightIcon aria-hidden="true" className="h-4" />
            </button>
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};

export const StoryBanners = ({ story }: { story: DetailedStory }) => {
  const [scheduleDialogNow, setScheduleDialogNow] = useState<number | null>(
    null,
  );
  const { getTermDisplay } = useTerminology();
  const { data: githubLinks = [] } = useStoryGitHubLinks(story.id);
  const { data: feedbackLinks = [] } = useStoryFeedbackLinks(story.id);
  const { data: requestLinks = [] } = useStoryIntegrationRequestLinks(story.id);
  const retryIssue = useRetryCalendarScheduleIssue();
  const overrideIssue = useOverrideCalendarScheduleIssue();
  const autoSchedulingStatus = deriveAutoSchedulingStatus(story);
  const items: BannerItem[] = [];

  if (isMayaBannerStatus(autoSchedulingStatus)) {
    const reason = getAutoSchedulingHelper(
      autoSchedulingStatus,
      story.autoSchedulingReason,
    ).replaceAll("story", getTermDisplay("storyTerm"));
    items.push({
      id: `maya:${autoSchedulingStatus}`,
      render: (embedded) => (
        <MayaScheduleBannerRow
          embedded={embedded}
          isRetrying={Boolean(
            retryIssue.isPending && retryIssue.variables === story.id,
          )}
          onChooseTime={
            autoSchedulingStatus === "cannot_fit" &&
            (story.estimatedDurationMinutes ?? 0) > 0
              ? () => {
                  setScheduleDialogNow(Date.now());
                }
              : undefined
          }
          onRetry={
            autoSchedulingStatus === "cannot_fit"
              ? () => {
                  retryIssue.mutate(story.id);
                }
              : undefined
          }
          reason={reason}
        />
      ),
    });
  }

  getGitHubBannerLinks(githubLinks).forEach((link) => {
    items.push({
      id: `github:${link.id}`,
      render: (embedded) => (
        <GitHubBannerRow embedded={embedded} link={link} storyId={story.id} />
      ),
    });
  });

  feedbackLinks.forEach((link: StoryFeedbackLink) => {
    items.push({
      id: `feedback:${link.id}`,
      render: (embedded) => (
        <FeedbackBannerRow embedded={embedded} link={link} />
      ),
    });
  });

  requestLinks.forEach((link: IntegrationRequestProviderThread) => {
    items.push({
      id: `request:${link.id}`,
      render: (embedded) => (
        <IntegrationRequestBannerRow embedded={embedded} link={link} />
      ),
    });
  });

  return (
    <>
      <StoryBannerStack items={items} />
      {scheduleDialogNow !== null ? (
        <ScheduleIssueDialog
          isSaving={overrideIssue.isPending}
          issue={{
            estimatedDurationMinutes: story.estimatedDurationMinutes,
            storyCode: `${story.teamCode}-${story.sequenceId}`,
            storyTitle: story.title,
          }}
          now={scheduleDialogNow}
          onClose={() => {
            setScheduleDialogNow(null);
          }}
          onSubmit={(startAt) => {
            overrideIssue.mutate(
              {
                storyId: story.id,
                startAt,
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
              },
              {
                onSuccess: () => {
                  setScheduleDialogNow(null);
                },
              },
            );
          }}
        />
      ) : null}
    </>
  );
};
