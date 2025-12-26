import {
  Tabs,
  Text,
  Flex,
  ProgressBar,
  Avatar,
  Container,
  Box,
  Divider,
  Button,
  DatePicker,
} from "ui";
import { useParams } from "next/navigation";
import { cn } from "lib";
import type { ReactNode } from "react";
import { CalendarIcon } from "icons";
import { format, formatISO } from "date-fns";
import { useSession } from "next-auth/react";
import { RowWrapper, StoryStatusIcon, PriorityIcon } from "@/components/ui";
import type { StoryPriority } from "@/modules/stories/types";
import {
  useObjective,
  useUpdateObjectiveMutation,
} from "@/modules/objectives/hooks";
import { useObjectiveAnalytics } from "@/modules/objectives/hooks/objective-analytics";
import type { ObjectiveUpdate } from "@/modules/objectives/types";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { ProgressChart } from "./progress-chart";

const Option = ({
  label,
  value,
  className,
}: {
  label: string;
  value: ReactNode;
  className?: string;
}) => {
  return (
    <Box
      className={cn(
        "my-3.5 grid grid-cols-[10rem_auto] items-center gap-3",
        className,
      )}
    >
      <Text
        className="flex items-center gap-1 truncate"
        color="muted"
        fontWeight="medium"
      >
        {label}
      </Text>
      {value}
    </Box>
  );
};

export const Sidebar = ({ className }: { className?: string }) => {
  const { data: session } = useSession();
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: objective } = useObjective(objectiveId);
  const { data: analytics } = useObjectiveAnalytics(objectiveId);
  const updateMutation = useUpdateObjectiveMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId,
      data,
    });
  };

  if (!analytics) {
    return null;
  }

  const {
    progressBreakdown,
    teamAllocation,
    priorityBreakdown,
    progressChart,
  } = analytics;
  const canUpdate = isAdminOrOwner || session?.user?.id === objective?.leadUser;

  // Map progress breakdown to status categories
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
    <Container
      className={cn(
        "bg-linear-to-br from-white via-gray-50/50 to-gray-50 py-6 dark:from-dark-200/50 dark:to-dark md:pl-8",
        className,
      )}
    >
      <Box className="mb-6">
        <Text fontWeight="semibold">Properties</Text>
        <Option
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <CalendarIcon
                      className={cn("h-[1.15rem] w-auto", {
                        "text-gray/80 dark:text-gray-300/80":
                          !objective?.startDate,
                      })}
                    />
                  }
                  variant="naked"
                >
                  {objective?.startDate ? (
                    format(new Date(objective.startDate), "MMM d, yy")
                  ) : (
                    <Text as="span" color="muted">
                      Start date
                    </Text>
                  )}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day: Date) => {
                  handleUpdate({
                    startDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective?.startDate
                    ? new Date(objective.startDate)
                    : undefined
                }
              />
            </DatePicker>
          }
        />
        <Option
          label="Target date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  className={cn({
                    "text-gray/80 dark:text-gray-300/80": !objective?.endDate,
                  })}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <CalendarIcon
                      className={cn("h-[1.15rem]", {
                        "text-gray/80 dark:text-gray-300/80":
                          !objective?.endDate,
                      })}
                    />
                  }
                  variant="naked"
                >
                  {objective?.endDate ? (
                    format(new Date(objective.endDate), "MMM d, yy")
                  ) : (
                    <Text color="muted">Target date</Text>
                  )}
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                onDayClick={(day: Date) => {
                  handleUpdate({
                    endDate: formatISO(day, { representation: "date" }),
                  });
                }}
                selected={
                  objective?.endDate ? new Date(objective.endDate) : undefined
                }
              />
            </DatePicker>
          }
        />
      </Box>
      <Divider className="my-6" />
      <Box>
        <Text>Progress Chart</Text>
        <ProgressChart progressData={progressChart} />
      </Box>
      <Divider className="my-6" />
      <Text className="mb-3">Stories Overview</Text>
      <Tabs defaultValue="assignees">
        <Tabs.List className="mx-0 mb-3 md:mx-0">
          <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
          <Tabs.Tab value="status">Status</Tabs.Tab>
          <Tabs.Tab value="priority">Priority</Tabs.Tab>
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
                <RowWrapper className="px-1 py-2 md:px-0" key={member.memberId}>
                  <Flex align="center" gap={2}>
                    <Avatar
                      name={member.username}
                      size="xs"
                      src={member.avatarUrl}
                    />
                    <Text color="muted">{member.username}</Text>
                  </Flex>
                  <Flex align="center" gap={2}>
                    <Text color="muted">
                      {member.completed} of {member.assigned}
                    </Text>
                    <ProgressBar className="w-20" progress={completionRate} />
                  </Flex>
                </RowWrapper>
              );
            })}
        </Tabs.Panel>

        <Tabs.Panel value="status">
          {Object.entries(progressBreakdown)
            .filter(([key, count]) => key !== "total" && count > 0)
            .map(([status, count]) => {
              const percentage =
                progressBreakdown.total > 0
                  ? (count / progressBreakdown.total) * 100
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
                    <Text color="muted">
                      {count} of {progressBreakdown.total}
                    </Text>
                    <ProgressBar className="w-20" progress={percentage} />
                  </Flex>
                </RowWrapper>
              );
            })}
        </Tabs.Panel>

        <Tabs.Panel value="priority">
          {priorityBreakdown
            .filter(({ count }) => count > 0)
            .map(({ priority, count }) => {
              const percentage =
                progressBreakdown.total > 0
                  ? (count / progressBreakdown.total) * 100
                  : 0;
              return (
                <RowWrapper className="px-1 py-2 md:px-0" key={priority}>
                  <Flex align="center" gap={2}>
                    <PriorityIcon priority={priority as StoryPriority} />
                    <Text color="muted">{priority}</Text>
                  </Flex>
                  <Flex align="center" gap={2}>
                    <Text color="muted">
                      {count} of {progressBreakdown.total}
                    </Text>
                    <ProgressBar className="w-20" progress={percentage} />
                  </Flex>
                </RowWrapper>
              );
            })}
        </Tabs.Panel>
      </Tabs>
    </Container>
  );
};
