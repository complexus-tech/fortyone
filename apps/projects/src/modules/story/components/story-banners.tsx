"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { formatISO } from "date-fns";
import {
  AiIcon,
  AssigneeIcon,
  CalendarIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  MoreHorizontalIcon,
  RefreshIcon,
  Time02Icon,
  TimeScheduleIcon,
} from "icons";
import { Box, Calendar, Flex, Menu, Text } from "ui";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { ScheduleIssueDialog } from "@/components/ui/story/schedule-issue-dialog";
import { SignalBannerRow } from "@/components/ui/signal-banner-row";
import { TimeNeededMenu } from "@/components/ui/story/time-needed-menu";
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
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import type { DetailedStory, StoryUpdate } from "../types";
import { FeedbackBannerRow } from "./feedback-banner";
import { GitHubBannerRow } from "./github-banner";
import { IntegrationRequestBannerRow } from "./integration-request-banner";

type BannerItem = {
  id: string;
  render: (embedded: boolean) => ReactNode;
};

const BANNER_ROTATION_INTERVAL_MS = 5000;

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

const MayaBannerActions = ({
  status,
  story,
}: {
  status: MayaBannerStatus;
  story: DetailedStory;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [scheduleDialogNow, setScheduleDialogNow] = useState<number | null>(
    null,
  );
  const retryIssue = useRetryCalendarScheduleIssue();
  const overrideIssue = useOverrideCalendarScheduleIssue();
  const updateStory = useUpdateStoryMutation();
  const isRetrying = Boolean(
    retryIssue.isPending && retryIssue.variables === story.id,
  );

  const updateSchedulingInput = (payload: StoryUpdate) => {
    updateStory.mutate({ payload, storyId: story.id });
    setIsOpen(false);
  };

  return (
    <>
      <Menu onOpenChange={setIsOpen} open={isOpen}>
        <Menu.Button>
          <button
            aria-label="More Maya scheduling actions"
            className="text-primary hover:text-primary/80 rounded-md p-1 transition"
            type="button"
          >
            <MoreHorizontalIcon className="h-5 text-current" />
          </button>
        </Menu.Button>
        <Menu.Items align="end" className="min-w-48">
          <Menu.Group>
            {status === "needs_time" ? (
              <Menu.SubMenu>
                <Menu.SubTrigger className="justify-between gap-4">
                  <Flex align="center" className="min-w-0 gap-1.5">
                    <Time02Icon className="h-5 w-auto shrink-0" />
                    <Text className="truncate">Add time needed</Text>
                  </Flex>
                  <ChevronRightIcon className="text-text-muted h-3.5 w-auto" />
                </Menu.SubTrigger>
                <TimeNeededMenu.SubItems
                  estimatedDurationMinutes={story.estimatedDurationMinutes}
                  minimumFocusBlockMinutes={story.minimumFocusBlockMinutes}
                  onDurationSelected={() => {
                    setIsOpen(false);
                  }}
                  setTimeNeeded={updateSchedulingInput}
                />
              </Menu.SubMenu>
            ) : null}

            {status === "needs_owner" ? (
              <AssigneesMenu.SubMenu
                assigneeId={story.assigneeId}
                disallowEmptySelection
                onAssigneeSelected={(assigneeId) => {
                  updateSchedulingInput({ assigneeId });
                }}
                teamId={story.teamId}
              >
                <Flex align="center" className="min-w-0 gap-1.5">
                  <AssigneeIcon className="h-5 w-auto shrink-0" />
                  <Text className="truncate">Choose owner</Text>
                </Flex>
                <ChevronRightIcon className="text-text-muted h-3.5 w-auto" />
              </AssigneesMenu.SubMenu>
            ) : null}

            {status === "at_risk" ? (
              <Menu.SubMenu>
                <Menu.SubTrigger className="justify-between gap-4">
                  <Flex align="center" className="min-w-0 gap-1.5">
                    <CalendarIcon className="h-5 w-auto shrink-0" />
                    <Text className="truncate">Update end date</Text>
                  </Flex>
                  <ChevronRightIcon className="text-text-muted h-3.5 w-auto" />
                </Menu.SubTrigger>
                <Menu.SubItems className="w-auto overflow-hidden p-0">
                  <Calendar
                    className="px-3 py-3 shadow-none"
                    mode="single"
                    onDayClick={(day) => {
                      updateSchedulingInput({
                        endDate: formatISO(day, { representation: "date" }),
                      });
                    }}
                    selected={
                      story.endDate ? new Date(story.endDate) : undefined
                    }
                  />
                </Menu.SubItems>
              </Menu.SubMenu>
            ) : null}

            {status === "cannot_fit" ? (
              <>
                <Menu.Item
                  disabled={isRetrying}
                  onSelect={() => {
                    retryIssue.mutate(story.id);
                  }}
                >
                  <RefreshIcon className="h-5 w-auto" />
                  {isRetrying ? "Retrying…" : "Retry"}
                </Menu.Item>
                <Menu.Item
                  disabled={(story.estimatedDurationMinutes ?? 0) <= 0}
                  onSelect={() => {
                    setScheduleDialogNow(Date.now());
                  }}
                >
                  <TimeScheduleIcon className="h-5 w-auto" />
                  Choose time
                </Menu.Item>
              </>
            ) : null}
          </Menu.Group>
        </Menu.Items>
      </Menu>

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

const MayaScheduleBannerRow = ({
  embedded,
  reason,
  status,
  story,
}: {
  embedded: boolean;
  reason: string;
  status: MayaBannerStatus;
  story: DetailedStory;
}) => (
  <SignalBannerRow
    actions={<MayaBannerActions status={status} story={story} />}
    icon={<AiIcon className="text-primary h-5 shrink-0" />}
    title={`Maya needs your help · ${reason}`}
    variant={embedded ? "embedded" : "standalone"}
  >
    Maya needs your help
    <span aria-hidden="true"> · </span>
    {reason}
  </SignalBannerRow>
);

const StoryBannerStack = ({ items }: { items: BannerItem[] }) => {
  const [activeIndex, setActiveIndex] = useState(0);
  const [isDocumentHidden, setIsDocumentHidden] = useState(false);
  const [isFocusWithin, setIsFocusWithin] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const [isRotationPaused, setIsRotationPaused] = useState(false);
  const isAutoRotating =
    items.length > 1 &&
    !isDocumentHidden &&
    !isFocusWithin &&
    !isHovered &&
    !isRotationPaused;

  useEffect(() => {
    const motionPreference = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    );
    const pauseForReducedMotion = () => {
      if (motionPreference.matches) setIsRotationPaused(true);
    };

    pauseForReducedMotion();
    motionPreference.addEventListener("change", pauseForReducedMotion);
    return () => {
      motionPreference.removeEventListener("change", pauseForReducedMotion);
    };
  }, []);

  useEffect(() => {
    const updateDocumentVisibility = () => {
      setIsDocumentHidden(document.hidden);
    };

    updateDocumentVisibility();
    document.addEventListener("visibilitychange", updateDocumentVisibility);
    return () => {
      document.removeEventListener(
        "visibilitychange",
        updateDocumentVisibility,
      );
    };
  }, []);

  useEffect(() => {
    if (!isAutoRotating) return;

    const timeoutId = window.setTimeout(() => {
      setActiveIndex((currentIndex) => (currentIndex + 1) % items.length);
    }, BANNER_ROTATION_INTERVAL_MS);

    return () => {
      window.clearTimeout(timeoutId);
    };
  }, [activeIndex, isAutoRotating, items.length]);

  if (items.length === 0) return null;

  const safeActiveIndex = activeIndex % items.length;
  const activeItem = items[safeActiveIndex];
  if (items.length === 1) {
    return <Box className="mb-3">{activeItem.render(false)}</Box>;
  }

  return (
    <Box
      aria-live={isAutoRotating ? "off" : "polite"}
      className="relative mb-3 pb-1.5"
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) {
          setIsFocusWithin(false);
        }
      }}
      onFocus={() => {
        setIsFocusWithin(true);
      }}
      onMouseEnter={() => {
        setIsHovered(true);
      }}
      onMouseLeave={() => {
        setIsFocusWithin(false);
        setIsHovered(false);
      }}
    >
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
          <Flex align="center" className="gap-1">
            <button
              aria-label={
                isRotationPaused
                  ? "Play banner rotation"
                  : "Pause banner rotation"
              }
              className="text-primary hover:text-primary/80 focus-visible:ring-ring mr-1 rounded-sm text-xs font-medium outline-none focus-visible:ring-2"
              onClick={() => {
                if (isRotationPaused) setIsFocusWithin(false);
                setIsRotationPaused(!isRotationPaused);
              }}
              type="button"
            >
              {isRotationPaused ? "Play" : "Pause"}
            </button>
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
  const { getTermDisplay } = useTerminology();
  const { data: githubLinks = [] } = useStoryGitHubLinks(story.id);
  const { data: feedbackLinks = [] } = useStoryFeedbackLinks(story.id);
  const { data: requestLinks = [] } = useStoryIntegrationRequestLinks(story.id);
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
          reason={reason}
          status={autoSchedulingStatus}
          story={story}
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

  return <StoryBannerStack items={items} />;
};
