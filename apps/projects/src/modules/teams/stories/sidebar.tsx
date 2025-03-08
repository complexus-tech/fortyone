import {
  Box,
  Tabs,
  Text,
  Flex,
  ProgressBar,
  Divider,
  Avatar,
  Tooltip,
  Badge,
} from "ui";
import { StoryIcon, SprintsIcon, TeamIcon, TagsIcon } from "icons";
import { useParams } from "next/navigation";
import { format, addDays, isAfter, isBefore, isToday } from "date-fns";
import {
  RowWrapper,
  StoryStatusIcon,
  PriorityIcon,
  TeamColor,
} from "@/components/ui";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import type { StoryPriority } from "@/modules/stories/types";
import { useLabels } from "@/lib/hooks/labels";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useTeams } from "../hooks/teams";

export const Sidebar = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: stories = [] } = useTeamStories(teamId);
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useTeamMembers(teamId);
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: sprints = [] } = useTeamSprints(teamId);
  const { data: labels = [] } = useLabels();

  const team = teams.find((t) => t.id === teamId)!;
  const teamMembers = members;
  const totalStories = stories.length;

  // Calculate story statistics
  const storiesByStatus = statuses.map((status) => {
    const count = stories.filter((s) => s.statusId === status.id).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { status, count, percentage };
  });

  const priorities: StoryPriority[] = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ];
  const storiesByPriority = priorities.map((priority) => {
    const count = stories.filter((s) => s.priority === priority).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { priority, count, percentage };
  });

  const storiesByAssignee = teamMembers.map((member) => {
    const count = stories.filter((s) => s.assigneeId === member.id).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { member, count, percentage };
  });

  const unassignedCount = stories.filter((s) => !s.assigneeId).length;

  const completedStories = stories.filter((s) => {
    const status = statuses.find((st) => st.id === s.statusId);
    return status?.category === "completed";
  }).length;

  const completionRate =
    totalStories > 0 ? (completedStories / totalStories) * 100 : 0;

  // Calculate sprint statistics
  const activeSprint = sprints.find(
    (sprint) =>
      isBefore(new Date(sprint.startDate), new Date()) &&
      isAfter(new Date(sprint.endDate), new Date()),
  );

  // Calculate due date statistics
  const storiesWithDueDate = stories.filter((s) => s.endDate);
  const overdueTasks = storiesWithDueDate.filter((s) =>
    isBefore(new Date(s.endDate!), new Date()),
  ).length;
  const dueTodayTasks = storiesWithDueDate.filter((s) =>
    isToday(new Date(s.endDate!)),
  ).length;
  const dueThisWeekTasks = storiesWithDueDate.filter(
    (s) =>
      isAfter(new Date(s.endDate!), new Date()) &&
      isBefore(new Date(s.endDate!), addDays(new Date(), 7)),
  ).length;

  // Calculate label statistics
  const allLabels = stories.reduce<Record<string, number>>((acc, story) => {
    story.labels.forEach((labelId) => {
      if (!acc[labelId]) {
        acc[labelId] = 0;
      }
      acc[labelId]++;
    });
    return acc;
  }, {});

  const labelStats = Object.entries(allLabels)
    .map(([labelId, count]) => {
      const label = labels.find((l) => l.id === labelId);
      return {
        label,
        count,
        percentage: (count / totalStories) * 100,
      };
    })
    .filter((stat) => stat.label); // Filter out any labels that don't exist anymore

  return (
    <Box className="py-6">
      <Flex align="center" className="mb-4 px-6" justify="between">
        <Text className="flex items-center gap-2">
          <StoryIcon className="h-5 w-auto" strokeWidth={2} />
          <span className="first-letter:uppercase">Team stories</span>
        </Text>
        <Text className="flex items-center gap-1.5">
          <TeamColor color={team.color} />
          <span
            className="inline-block max-w-[16ch] truncate"
            title={team.name}
          >
            {team.name}
          </span>
        </Text>
      </Flex>

      <Divider className="mb-4" />

      {/* Team Details Section */}
      <Box className="mb-4 px-6">
        <Text className="mb-3">Team Details</Text>
        <Flex direction="column" gap={3}>
          <Flex align="center" justify="between">
            <Text color="muted">Team Code</Text>
            <Text color="muted">{team.code}</Text>
          </Flex>
          <Flex align="center" justify="between">
            <Text color="muted">Members</Text>
            <Flex align="center" gap={2}>
              <TeamIcon className="h-4" />
              <Text color="muted">{teamMembers.length}</Text>
            </Flex>
          </Flex>
          <Flex align="center" justify="between">
            <Text color="muted">Created</Text>
            <Text color="muted">
              {format(new Date(team.createdAt), "MMM d, yyyy")}
            </Text>
          </Flex>
          <Flex align="center" justify="between">
            <Text color="muted">Completion Rate</Text>
            <Flex align="center" gap={2}>
              <ProgressBar className="w-20" progress={completionRate} />
              <Text color="muted">{completionRate.toFixed(1)}%</Text>
            </Flex>
          </Flex>
        </Flex>
      </Box>

      <Divider className="mb-4" />

      {/* Stories Overview Section */}
      <Box className="px-6">
        <Text className="mb-3">Stories Overview</Text>
        <Tabs defaultValue="status">
          <Tabs.List className="mx-0 mb-3">
            <Tabs.Tab value="status">Status</Tabs.Tab>
            <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
            <Tabs.Tab value="priority">Priority</Tabs.Tab>
            <Tabs.Tab value="labels">Labels</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="status">
            {storiesByStatus.map(({ status, count }) => (
              <RowWrapper
                className="px-1 py-2 last-of-type:border-0 md:px-0"
                key={status.id}
              >
                <Flex align="center" gap={2}>
                  <StoryStatusIcon statusId={status.id} />
                  <Text color="muted">{status.name}</Text>
                </Flex>
                <Text color="muted">
                  {count} of {totalStories}
                </Text>
              </RowWrapper>
            ))}
          </Tabs.Panel>

          <Tabs.Panel value="assignees">
            {storiesByAssignee.map(({ member, count }) => (
              <RowWrapper
                className="px-1 py-2 last-of-type:border-0 md:px-0"
                key={member.id}
              >
                <Flex align="center" gap={2}>
                  <Avatar
                    className="h-6"
                    name={member.fullName}
                    src={member.avatarUrl}
                  />
                  <Text color="muted">{member.username}</Text>
                </Flex>
                <Text color="muted">
                  {count} of {totalStories}
                </Text>
              </RowWrapper>
            ))}
            <RowWrapper className="border-b-0 px-1 py-2 md:px-0">
              <Flex align="center" gap={2}>
                <Avatar className="h-6" />
                <Text color="muted">Unassigned</Text>
              </Flex>
              <Text color="muted">
                {unassignedCount} of {totalStories}
              </Text>
            </RowWrapper>
          </Tabs.Panel>

          <Tabs.Panel value="priority">
            {storiesByPriority.map(({ priority, count }) => (
              <RowWrapper
                className="px-1 py-2 last-of-type:border-0 md:px-0"
                key={priority}
              >
                <Flex align="center" gap={2}>
                  <PriorityIcon priority={priority} />
                  <Text color="muted">{priority}</Text>
                </Flex>
                <Text color="muted">
                  {count} of {totalStories}
                </Text>
              </RowWrapper>
            ))}
          </Tabs.Panel>

          <Tabs.Panel value="labels">
            {labelStats.map(({ label, count }) => (
              <RowWrapper
                className="px-1 py-2 last-of-type:border-0 md:px-0"
                key={label!.id}
              >
                <Flex align="center" gap={2}>
                  <TagsIcon className="h-4" style={{ color: label!.color }} />
                  <Text color="muted">{label!.name}</Text>
                </Flex>
                <Text color="muted">
                  {count} of {totalStories}
                </Text>
              </RowWrapper>
            ))}
          </Tabs.Panel>
        </Tabs>
      </Box>

      {/* Active Sprint Section */}
      {activeSprint ? (
        <>
          <Divider className="mb-4 mt-4" />
          <Box className="mb-4 px-6">
            <Text className="mb-4">Active Sprint</Text>
            <Flex direction="column" gap={3}>
              <Flex align="center" justify="between">
                <Flex align="center" gap={2}>
                  <SprintsIcon />
                  <Text>{activeSprint.name}</Text>
                </Flex>
                <Badge
                  className="h-8 px-2 text-base capitalize tracking-wide"
                  color="primary"
                >
                  In Progress
                </Badge>
              </Flex>
              <Flex align="center" justify="between">
                <Text color="muted">Sprint Progress</Text>
                <Text>
                  {(
                    (activeSprint.stats.completed / activeSprint.stats.total) *
                    100
                  ).toFixed(1)}
                  %
                </Text>
              </Flex>
              <ProgressBar
                progress={
                  (activeSprint.stats.completed / activeSprint.stats.total) *
                  100
                }
              />
              <Flex className="mt-2" gap={4}>
                <Tooltip title="Completed Stories">
                  <Flex align="center" gap={1}>
                    <StoryStatusIcon
                      statusId={
                        statuses.find((s) => s.category === "completed")?.id
                      }
                    />
                    <Text color="muted">{activeSprint.stats.completed}</Text>
                  </Flex>
                </Tooltip>
                <Tooltip title="In Progress Stories">
                  <Flex align="center" gap={1}>
                    <StoryStatusIcon
                      statusId={
                        statuses.find((s) => s.category === "started")?.id
                      }
                    />
                    <Text color="muted">{activeSprint.stats.started}</Text>
                  </Flex>
                </Tooltip>
                <Tooltip title="Todo Stories">
                  <Flex align="center" gap={1}>
                    <StoryStatusIcon
                      statusId={
                        statuses.find((s) => s.category === "unstarted")?.id
                      }
                    />
                    <Text color="muted">{activeSprint.stats.unstarted}</Text>
                  </Flex>
                </Tooltip>
              </Flex>
            </Flex>
          </Box>
        </>
      ) : null}

      {/* Due Dates Section */}
      <Divider className="mb-4" />
      <Box className="mb-4 px-6">
        <Text className="mb-4">Due Dates</Text>
        <Flex direction="column" gap={3}>
          <Flex align="center" justify="between">
            <Text color="muted">Overdue</Text>
            <Text color="muted">{overdueTasks}</Text>
          </Flex>
          <Flex align="center" justify="between">
            <Text color="muted">Due Today</Text>
            <Text color="muted">{dueTodayTasks}</Text>
          </Flex>
          <Flex align="center" justify="between">
            <Text color="muted">Due This Week</Text>
            <Text color="muted">{dueThisWeekTasks}</Text>
          </Flex>
        </Flex>
      </Box>
    </Box>
  );
};
