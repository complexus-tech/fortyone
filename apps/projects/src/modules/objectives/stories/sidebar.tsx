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
import { CalendarIcon, TagsIcon } from "icons";
import { format } from "date-fns";
import { useSession } from "next-auth/react";
import {
  RowWrapper,
  StoryStatusIcon,
  PriorityIcon,
  PrioritiesMenu,
  AssigneesMenu,
} from "@/components/ui";
import { useObjectiveStories } from "@/modules/stories/hooks/objective-stories";
import { useLabels } from "@/lib/hooks/labels";
import type { StoryPriority } from "@/modules/stories/types";
import {
  useObjective,
  useUpdateObjectiveMutation,
} from "@/modules/objectives/hooks";
import type { ObjectiveUpdate } from "@/modules/objectives/types";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { ObjectiveStatusesMenu } from "../../../components/ui/objective-statuses-menu";

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
  const { data: stories = [] } = useObjectiveStories(objectiveId);
  const { data: members = [] } = useTeamMembers(objective?.teamId);
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: labels = [] } = useLabels();
  const updateMutation = useUpdateObjectiveMutation();

  const handleUpdate = (data: ObjectiveUpdate) => {
    updateMutation.mutate({
      objectiveId,
      data,
    });
  };

  const totalStories = stories.length;
  const status = statuses.find((s) => s.id === objective?.statusId);
  const leadUser = members.find((m) => m.id === objective?.leadUser);
  const createdBy = members.find((m) => m.id === objective?.createdBy);
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const canUpdate = isAdminOrOwner || session?.user?.id === objective?.leadUser;
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

  const storiesByAssignee = members.map((member) => {
    const count = stories.filter((s) => s.assigneeId === member.id).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { member, count, percentage };
  });
  const storiesWithNoAssignee = stories.filter((s) => !s.assigneeId).length;

  const allLabels = stories.reduce<Record<string, number>>((acc, story) => {
    story.labels?.forEach((labelId) => {
      if (!acc[labelId]) acc[labelId] = 0;
      acc[labelId]++;
    });
    return acc;
  }, {});

  const labelStats = Object.entries(allLabels)
    .map(([labelId, count]) => {
      const label = labels.find((l) => l.id === labelId);
      const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
      return { label, count, percentage };
    })
    .filter((stat) => stat.label);

  return (
    <Container
      className={cn(
        "bg-gradient-to-br from-white via-gray-50/50 to-gray-50 py-6 dark:from-dark-200/50 dark:to-dark md:pl-8",
        className,
      )}
    >
      <Box className="mb-6">
        <Text fontWeight="semibold">Properties</Text>
        <Option
          label="Created by"
          value={
            <Button
              className="font-medium"
              color="tertiary"
              href={`/profile/${createdBy?.id}`}
              leftIcon={
                <Avatar
                  name={createdBy?.fullName || createdBy?.username}
                  size="xs"
                  src={createdBy?.avatarUrl}
                />
              }
              type="button"
              variant="naked"
            >
              {createdBy?.username}
            </Button>
          }
        />
        <Option
          label="Status"
          value={
            <ObjectiveStatusesMenu>
              <ObjectiveStatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <ObjectiveStatusIcon statusId={objective?.statusId} />
                  }
                  type="button"
                  variant="naked"
                >
                  {status?.name ?? "Backlog"}
                </Button>
              </ObjectiveStatusesMenu.Trigger>
              <ObjectiveStatusesMenu.Items
                setStatusId={(statusId) => {
                  handleUpdate({ statusId });
                }}
                statusId={objective?.statusId}
              />
            </ObjectiveStatusesMenu>
          }
        />
        <Option
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={<PriorityIcon priority={objective?.priority} />}
                  type="button"
                  variant="naked"
                >
                  {objective?.priority ?? "No Priority"}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items
                priority={objective?.priority}
                setPriority={(priority) => {
                  handleUpdate({ priority });
                }}
              />
            </PrioritiesMenu>
          }
        />
        <Option
          label="Lead"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  className={cn("font-medium", {
                    "text-gray-200 dark:text-gray-300": !objective?.leadUser,
                  })}
                  color="tertiary"
                  disabled={!canUpdate}
                  leftIcon={
                    <Avatar
                      className={cn({
                        "text-dark/80 dark:text-gray-200": !objective?.leadUser,
                      })}
                      name={leadUser?.username}
                      size="xs"
                      src={leadUser?.avatarUrl}
                    />
                  }
                  type="button"
                  variant="naked"
                >
                  {leadUser ? (
                    leadUser.username
                  ) : (
                    <Text as="span" color="muted">
                      Assign lead
                    </Text>
                  )}
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                assigneeId={objective?.leadUser}
                onAssigneeSelected={(leadUser) => {
                  handleUpdate({ leadUser: leadUser ?? undefined });
                }}
              />
            </AssigneesMenu>
          }
        />

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
                  handleUpdate({ startDate: day.toISOString() });
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
                  handleUpdate({ endDate: day.toISOString() });
                }}
                selected={
                  objective?.endDate ? new Date(objective.endDate) : undefined
                }
              />
            </DatePicker>
          }
        />
      </Box>
      <Divider className="mb-3" />
      <Text className="mb-3">Stories Overview</Text>
      <Tabs defaultValue="status">
        <Tabs.List className="mx-0 mb-3">
          <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
          <Tabs.Tab value="status">Status</Tabs.Tab>
          <Tabs.Tab value="labels">Labels</Tabs.Tab>
          <Tabs.Tab value="priority">Priority</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="assignees">
          {storiesByAssignee
            .filter((stat) => stat.count > 0)
            .map(({ member, count, percentage }) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={member.id}>
                <Flex align="center" gap={2}>
                  <Avatar
                    name={member.fullName}
                    size="xs"
                    src={member.avatarUrl}
                  />
                  <Text color="muted">{member.username}</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar className="w-20" progress={percentage} />
                  <Text color="muted">
                    {count} of {totalStories}
                  </Text>
                </Flex>
              </RowWrapper>
            ))}
          {storiesWithNoAssignee > 0 && (
            <RowWrapper className="px-1 py-2 md:px-0">
              <Flex align="center" gap={2}>
                <Avatar size="xs" />
                <Text color="muted">No Assignee</Text>
              </Flex>
              <Flex align="center" gap={2}>
                <ProgressBar
                  className="w-20"
                  progress={(storiesWithNoAssignee / totalStories) * 100}
                />
                <Text color="muted">
                  {storiesWithNoAssignee} of {totalStories}
                </Text>
              </Flex>
            </RowWrapper>
          )}
        </Tabs.Panel>
        <Tabs.Panel value="status">
          {storiesByStatus
            .filter((stat) => stat.count > 0)
            .map(({ status, count, percentage }) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={status.id}>
                <Flex align="center" gap={2}>
                  <StoryStatusIcon statusId={status.id} />
                  <Text color="muted">{status.name}</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar className="w-20" progress={percentage} />
                  <Text color="muted">
                    {count} of {totalStories}
                  </Text>
                </Flex>
              </RowWrapper>
            ))}
        </Tabs.Panel>
        <Tabs.Panel value="labels">
          {labelStats.map(({ label, count, percentage }) => (
            <RowWrapper className="px-1 py-2 md:px-0" key={label!.id}>
              <Flex align="center" gap={2}>
                <TagsIcon
                  className="h-4 w-auto"
                  style={{ color: label!.color }}
                />
                <Text color="muted">{label!.name}</Text>
              </Flex>
              <Flex align="center" gap={2}>
                <ProgressBar className="w-20" progress={percentage} />
                <Text color="muted">
                  {count} of {totalStories}
                </Text>
              </Flex>
            </RowWrapper>
          ))}
        </Tabs.Panel>
        <Tabs.Panel value="priority">
          {storiesByPriority
            .filter((stat) => stat.count > 0)
            .map(({ priority, count, percentage }) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={priority}>
                <Flex align="center" gap={2}>
                  <PriorityIcon priority={priority} />
                  <Text color="muted">{priority}</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar className="w-20" progress={percentage} />
                  <Text color="muted">
                    {count} of {totalStories}
                  </Text>
                </Flex>
              </RowWrapper>
            ))}
        </Tabs.Panel>
      </Tabs>
    </Container>
  );
};
