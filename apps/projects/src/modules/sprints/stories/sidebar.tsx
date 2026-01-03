import { Avatar, Badge, Box, Divider, Flex, ProgressBar, Tabs, Text } from "ui";
import { SprintsIcon } from "icons";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { RowWrapper, StoryStatusIcon } from "@/components/ui";
import { useSprint } from "../hooks/sprint-details";
import { useSprintAnalytics } from "../hooks/sprint-analytics";
import type { SprintAnalytics } from "../types";
import { BurndownChart } from "./burndown";

export const Sidebar = () => {
  const { sprintId } = useParams<{ sprintId: string }>();
  const { data: sprint, isPending: isSprintPending } = useSprint(sprintId);
  const { data: analytics, isPending: isAnalyticsPending } =
    useSprintAnalytics(sprintId);

  if (isSprintPending || isAnalyticsPending || !sprint || !analytics) {
    return null;
  }

  const { overview, storyBreakdown, teamAllocation } = analytics;

  // Get status color based on analytics status
  const getStatusColor = (status: SprintAnalytics["overview"]["status"]) => {
    switch (status) {
      case "on_track":
        return "success";
      case "behind":
        return "danger";
      case "completed":
        return "tertiary";
      case "at_risk":
        return "warning";
      case "not_started":
        return "tertiary";
      default:
        return "tertiary";
    }
  };

  const breakdownStatusMap = {
    inProgress: {
      label: "In Progress",
      category: "started",
    },
    todo: {
      label: "To Do",
      category: "unstarted",
    },
    blocked: {
      label: "Blocked",
      category: "paused",
    },
    cancelled: {
      label: "Cancelled",
      category: "cancelled",
    },
    completed: {
      label: "Completed",
      category: "completed",
    },
  } as const;

  return (
    <Box className="h-full bg-surface-muted/30 py-6 bg-surface/60">
      <Box className="px-6">
        <Flex align="center" justify="between">
          <Text className="flex items-center gap-1.5" fontSize="lg">
            <SprintsIcon className="relative -top-px h-[1.4rem] w-auto" />
            {sprint.name}
          </Text>
          {/* <Menu>
            <Menu.Button>
              <Button
                asIcon
                color="tertiary"
                leftIcon={<MoreHorizontalIcon />}
                rounded="full"
                size="sm"
                variant="naked"
              >
                <span className="sr-only">More options</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="w-48">
              <Menu.Group>
                <Menu.Item>
                  <EditIcon />
                  Edit {getTermDisplay("sprintTerm")}
                </Menu.Item>
                <Menu.Item>
                  <CopyIcon />
                  Copy link
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu> */}
        </Flex>
        <Flex align="center" className="my-4" gap={2}>
          <Badge
            className="h-8 px-2 text-base capitalize tracking-wide"
            color={getStatusColor(overview.status)}
          >
            {overview.status.replace("_", " ")}
          </Badge>
          <Badge
            className="h-8 px-2 text-base capitalize tracking-wide"
            color="tertiary"
          >
            {format(new Date(sprint.startDate), "d MMM")} -{" "}
            {format(new Date(sprint.endDate), "d MMM")}
          </Badge>
        </Flex>
        <Flex align="center" className="mb-2 mt-3" gap={2} justify="between">
          <Text>Sprint Progress</Text>
          <Text>{overview.completionPercentage}%</Text>
        </Flex>
        <ProgressBar className="h-2" progress={overview.completionPercentage} />
        <Flex align="center" className="mt-2" gap={4} justify="between">
          <Text color="muted">{overview.daysElapsed} days elapsed</Text>
          {overview.daysRemaining > 0 && (
            <Text color="muted">{overview.daysRemaining} days remaining</Text>
          )}
        </Flex>
      </Box>
      <Divider className="my-6" />
      <Box className="px-6">
        <Text>Burndown Chart</Text>
        <BurndownChart burndownData={analytics.burndown} />
      </Box>
      <Divider className="my-6" />
      <Box className="px-6">
        <Text className="mb-3">Stories Overview</Text>
        <Tabs defaultValue="assignees">
          <Tabs.List className="mx-0 mb-3 md:mx-0">
            <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
            <Tabs.Tab value="status">Status</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="assignees">
            {teamAllocation
              .filter((member) => member.assigned > 0 || member.completed > 0)
              .map((member) => {
                const completionRate =
                  member.assigned > 0
                    ? (member.completed / member.assigned) * 100
                    : 0;
                return (
                  <RowWrapper
                    className="px-1 py-2 md:px-0"
                    key={member.memberId}
                  >
                    <Flex align="center" gap={2}>
                      <Avatar
                        name={member.username}
                        size="xs"
                        src={member.avatarUrl}
                      />
                      <Text color="muted">{member.username}</Text>
                    </Flex>
                    <Flex align="center" gap={2}>
                      <ProgressBar className="w-20" progress={completionRate} />
                      <Text color="muted">
                        {member.completed} of {member.assigned}
                      </Text>
                    </Flex>
                  </RowWrapper>
                );
              })}
          </Tabs.Panel>

          <Tabs.Panel value="status">
            {Object.entries(storyBreakdown)
              .filter(([key, count]) => key !== "total" && count > 0)
              .map(([status, count]) => {
                const percentage =
                  storyBreakdown.total > 0
                    ? (count / storyBreakdown.total) * 100
                    : 0;
                const statusConfig =
                  breakdownStatusMap[status as keyof typeof breakdownStatusMap];
                const displayName = statusConfig.label;
                const category = statusConfig.category;

                return (
                  <RowWrapper className="px-1 py-2 md:px-0" key={status}>
                    <Flex align="center" gap={2}>
                      <StoryStatusIcon category={category} />
                      <Text color="muted">{displayName}</Text>
                    </Flex>
                    <Flex align="center" gap={2}>
                      <ProgressBar className="w-20" progress={percentage} />
                      <Text color="muted">
                        {count} of {storyBreakdown.total}
                      </Text>
                    </Flex>
                  </RowWrapper>
                );
              })}
          </Tabs.Panel>
        </Tabs>
      </Box>
    </Box>
  );
};
