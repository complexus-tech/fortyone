"use client";

import { Box, Flex, Text, Tooltip, Avatar, DatePicker } from "ui";
import { differenceInDays, formatISO } from "date-fns";
import { useRouter } from "next/navigation";
import { useCallback, useState } from "react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import type { DateRange } from "react-day-picker";
import { CalendarPlusIcon } from "icons";
import type { Story } from "@/modules/stories/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { slugify } from "@/utils";
import { storyKeys } from "@/modules/stories/constants";
import { getStory } from "@/modules/story/queries/get-story";
import { getStoryAttachments } from "@/modules/story/queries/get-attachments";
import { linkKeys } from "@/constants/keys";
import { getLinks } from "@/lib/queries/links/get-links";
import { PrioritiesMenu } from "@/components/ui/story/priorities-menu";
import { StatusesMenu } from "@/components/ui/story/statuses-menu";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { StoryContextMenu } from "@/components/ui/story/context-menu";
import type { DetailedStory } from "@/modules/story/types";
import { PriorityIcon } from "./priority-icon";
import { StoryStatusIcon } from "./story-status-icon";
import { BaseGantt, GanttHeader, type ZoomLevel } from "./base-gantt";

// Individual Story Row Component
const StoryRow = ({
  story,
  duration,
  getTeamCode,
  handleUpdate,
}: {
  story: Story;
  duration: number | null;
  getTeamCode: (teamId: string) => string;
  handleUpdate: (storyId: string, data: Partial<DetailedStory>) => void;
}) => {
  // Import router and userRole directly in this component
  const router = useRouter();
  const { userRole } = useUserRole();
  const { data: session } = useSession();
  const queryClient = useQueryClient();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const [dates, setDates] = useState<DateRange | undefined>(undefined);
  // Get team members for this specific story's team
  const { data: members = [] } = useTeamMembers(story.teamId);

  const selectedAssignee = members.find(
    (member) => member.id === story.assigneeId!,
  );

  return (
    <Box
      onMouseEnter={() => {
        if (session) {
          const ctx = { session, workspaceSlug };
          queryClient.prefetchQuery({
            queryKey: storyKeys.detail(workspaceSlug, story.id),
            queryFn: () => getStory(story.id, ctx),
          });
          queryClient.prefetchQuery({
            queryKey: storyKeys.attachments(workspaceSlug, story.id),
            queryFn: () => getStoryAttachments(story.id, ctx),
          });
          queryClient.prefetchQuery({
            queryKey: linkKeys.story(story.id),
            queryFn: () => getLinks(story.id, ctx),
          });
        }
        router.prefetch(withWorkspace(`/story/${story.id}/${slugify(story.title)}`));
      }}
    >
      <StoryContextMenu story={story}>
        <Flex
          align="center"
          className="group h-14 border-b-[0.5px] border-border px-6 transition-colors hover:bg-state-hover"
          justify="between"
        >
          <Flex align="center" className="min-w-0 flex-1 gap-2">
            <Text
              className="line-clamp-1 w-[4.1rem] shrink-0 text-[0.95rem]"
              color="muted"
            >
              {getTeamCode(story.teamId)}-{story.sequenceId}
            </Text>
            <AssigneesMenu>
              <Tooltip
                className="py-2.5"
                title={
                  selectedAssignee ? (
                    <Box>
                      <Flex gap={2}>
                        <Avatar
                          className="mt-0.5"
                          name={selectedAssignee.fullName}
                          size="sm"
                          src={selectedAssignee.avatarUrl}
                        />
                        <Box>
                          <Text fontSize="md" fontWeight="medium">
                            {selectedAssignee.fullName}
                          </Text>
                          <Text color="muted" fontSize="md">
                            ({selectedAssignee.username})
                          </Text>
                        </Box>
                      </Flex>
                    </Box>
                  ) : null
                }
              >
                <span>
                  <AssigneesMenu.Trigger>
                    <button
                      className="flex"
                      disabled={userRole === "guest"}
                      type="button"
                    >
                      <Avatar
                        name={
                          selectedAssignee?.fullName ||
                          selectedAssignee?.username
                        }
                        size="xs"
                        src={selectedAssignee?.avatarUrl}
                      />
                    </button>
                  </AssigneesMenu.Trigger>
                </span>
              </Tooltip>
              <AssigneesMenu.Items
                assigneeId={selectedAssignee?.id}
                onAssigneeSelected={(assigneeId) => {
                  handleUpdate(story.id, { assigneeId });
                }}
                teamId={story.teamId}
              />
            </AssigneesMenu>

            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <button
                  className="flex shrink-0 select-none items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={userRole === "guest"}
                  type="button"
                >
                  <PriorityIcon priority={story.priority} />
                  <span className="sr-only">{story.priority}</span>
                </button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={story.priority}
                setPriority={(priority) => {
                  handleUpdate(story.id, { priority });
                }}
              />
            </PrioritiesMenu>

            <StatusesMenu>
              <StatusesMenu.Trigger>
                <button
                  className="flex items-center gap-1 disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={userRole === "guest"}
                  type="button"
                >
                  <StoryStatusIcon statusId={story.statusId} />
                  <span className="sr-only">Story status</span>
                </button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items
                setStatusId={(statusId) => {
                  handleUpdate(story.id, { statusId });
                }}
                statusId={story.statusId}
                teamId={story.teamId}
              />
            </StatusesMenu>

            <Link
              className="flex min-w-0 flex-1 items-center gap-1.5"
              href={withWorkspace(`/story/${story.id}/${slugify(story.title)}`)}
            >
              <Text
                className="line-clamp-1 hover:opacity-90"
                fontWeight="medium"
              >
                {story.title}
              </Text>
            </Link>
          </Flex>

          {duration ? (
            <Text className="ml-4 shrink-0" color="muted">
              {duration} day{duration !== 1 ? "s" : ""}
            </Text>
          ) : (
            <DatePicker>
              <Tooltip title="Add dates">
                <span className="mt-1">
                  <DatePicker.Trigger>
                    <button type="button">
                      <CalendarPlusIcon />
                    </button>
                  </DatePicker.Trigger>
                </span>
              </Tooltip>
              <DatePicker.Calendar
                mode="range"
                numberOfMonths={2}
                onSelect={(range) => {
                  setDates(range);
                  if (range?.from && range.to) {
                    handleUpdate(story.id, {
                      startDate: formatISO(range.from, {
                        representation: "date",
                      }),
                      endDate: formatISO(range.to, { representation: "date" }),
                    });
                  }
                }}
                selected={dates}
              />
            </DatePicker>
          )}
        </Flex>
      </StoryContextMenu>
    </Box>
  );
};

type GanttBoardProps = {
  stories: Story[];
  className?: string;
};

export const GanttBoard = ({ stories, className }: GanttBoardProps) => {
  const { data: teams = [] } = useTeams();
  const { mutate } = useUpdateStoryMutation();
  const router = useRouter();
  const { withWorkspace } = useWorkspacePath();
  
  // Simple function to get team code from teamId
  const getTeamCode = (teamId: string): string => {
    const team = teams.find((t) => t.id === teamId);
    return team?.code || "STORY";
  };

  // Handle date updates from drag operations
  const handleDateUpdate = useCallback(
    (storyId: string, startDate: string, endDate: string) => {
      mutate({
        storyId,
        payload: {
          startDate,
          endDate,
        },
      });
    },
    [mutate],
  );

  // Handle bar clicks to navigate to story page
  const handleBarClick = useCallback(
    (story: Story) => {
      router.push(withWorkspace(`/story/${story.id}/${slugify(story.title)}`));
    },
    [router],
  );

  const handleUpdate = useCallback(
    (storyId: string, data: Partial<DetailedStory>) => {
      mutate({
        storyId,
        payload: data,
      });
    },
    [mutate],
  );

  // Render sidebar for stories
  const renderSidebar = useCallback(
    (
      stories: Story[],
      onReset: () => void,
      zoomLevel: ZoomLevel,
      onZoomChange: (zoom: ZoomLevel) => void,
    ) => {
      return (
        <Box className="sticky left-0 z-20 w-screen shrink-0 border-r-[0.5px] border-border/60 bg-background md:w-136">
          <GanttHeader
            onReset={onReset}
            onZoomChange={onZoomChange}
            zoomLevel={zoomLevel}
          />
          {stories.map((story) => {
            const startDate = story.startDate
              ? new Date(story.startDate)
              : null;
            const endDate = story.endDate ? new Date(story.endDate) : null;
            const duration =
              startDate && endDate
                ? differenceInDays(endDate, startDate)
                : null;

            return (
              <StoryRow
                duration={duration}
                getTeamCode={getTeamCode}
                handleUpdate={handleUpdate}
                key={story.id}
                story={story}
              />
            );
          })}
        </Box>
      );
    },
    [getTeamCode, handleUpdate],
  );

  // Render bar content
  const renderBarContent = useCallback(
    (story: Story) => (
      <Text className="line-clamp-1" fontWeight="medium">
        {story.title}
      </Text>
    ),
    [],
  );

  return (
    <BaseGantt
      className={className}
      items={stories}
      onBarClick={handleBarClick}
      onDateUpdate={handleDateUpdate}
      renderBarContent={renderBarContent}
      renderSidebar={renderSidebar}
      storageKey="zoomLevel"
    />
  );
};
